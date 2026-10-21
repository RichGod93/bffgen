// HTTP checks for generated routes against mock.js. Needs a project generated with the sample
// backends from fixtures.py, running with JWT_SECRET=testsecret. run.sh does all of that.
// usage: node checks.js <bff-base-url>
const crypto = require('crypto');
const [base = 'http://localhost:8181'] = process.argv.slice(2);
const b64 = obj => Buffer.from(JSON.stringify(obj)).toString('base64url');
const sign = claims => {
  const unsigned = `${b64({ alg: 'HS256', typ: 'JWT' })}.${b64({ userId: 'u1', sub: 'u1', exp: Math.floor(Date.now() / 1000) + 600, ...claims })}`;
  return `${unsigned}.${crypto.createHmac('sha256', 'testsecret').update(unsigned).digest('base64url')}`;
};
const token = sign({ role: 'user' });
const auth = { authorization: `Bearer ${token}` };
const bearer = t => ({ headers: { authorization: `Bearer ${t}` } });

let failed = 0;
const check = (name, ok, detail = '') => {
  console.log(`${ok ? 'PASS' : 'FAIL'}  ${name}${ok ? '' : `  -> ${detail}`}`);
  if (!ok) failed++;
};
const call = async (path, init = {}) => {
  const started = Date.now();
  const res = await fetch(base + path, { ...init, headers: { ...auth, ...init.headers } });
  const text = await res.text();
  let body = null;
  try { body = JSON.parse(text); } catch { body = text; }
  return { status: res.status, headers: res.headers, body, ms: Date.now() - started };
};
const show = r => JSON.stringify({ status: r.status, body: r.body });

(async () => {
  let r = await call('/api/users/42?limit=5&tag=a&tag=b', { headers: { 'x-request-id': 'trace-1' } });
  check('path param substituted, query forwarded verbatim', r.status === 200 && r.body.url === '/api/users/42?limit=5&tag=a&tag=b', show(r));
  check('authorization and x-request-id forwarded', r.body.auth === true && r.body.requestId === 'trace-1', show(r));
  check('allowlisted response header forwarded (cache-control)', r.headers.get('cache-control') === 'max-age=5', r.headers.get('cache-control'));
  check('other backend headers not forwarded (x-internal)', r.headers.get('x-internal') === null, r.headers.get('x-internal'));

  r = await call('/api/users/a%20b%3Fc');
  check('param value re-encoded', r.body.url === '/api/users/a%20b%3Fc', show(r));

  // Starlette (FastAPI) decodes %2F before routing, so an encoded slash can't be part of a param there
  if (!process.env.NO_ENCODED_SLASH) {
    r = await call('/api/users/a%2Fb');
    check('encoded slash kept in param', r.body.url === '/api/users/a%2Fb', show(r));
  }

  r = await call('/api/orders/7/items/9');
  check('{param} style, two params', r.body.url === '/api/orders/7/items/9', show(r));

  r = await call('/api/users', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ name: 'x' }) });
  check('POST body forwarded', r.status === 200 && r.body.method === 'POST' && r.body.body === '{"name":"x"}', show(r));

  r = await call('/api/users/42', { method: 'HEAD' });
  check('HEAD answered without body', r.status === 200 && r.body === '', show(r));

  r = await call('/api/users/missing');
  check('backend 404 passed through', r.status === 404 && r.body.error === 'not found', show(r));

  r = await call('/api/users/empty');
  check('backend 204 passed through', r.status === 204 && r.body === '', show(r));

  r = await call('/api/users/42', { headers: { authorization: '' } });
  check('missing token -> 401', r.status === 401, show(r));
  check('401 has a WWW-Authenticate: Bearer challenge (RFC 9110)', /^Bearer\b/.test(r.headers.get('www-authenticate') || ''), r.headers.get('www-authenticate'));
  r = await call('/api/users/42', bearer(sign({ nbf: Math.floor(Date.now() / 1000) + 600 })));
  check('not-yet-valid token (nbf in the future) -> 401', r.status === 401, show(r));

  // tampered: payload changed after signing, so the signature no longer matches
  const [head, , sig] = token.split('.');
  r = await call('/api/users/42', bearer(`${head}.${b64({ userId: 'u1', sub: 'u1', role: 'admin', exp: Math.floor(Date.now() / 1000) + 600 })}.${sig}`));
  check('tampered token -> 401', r.status === 401, show(r));
  r = await call('/api/users/42', bearer(sign({ exp: Math.floor(Date.now() / 1000) - 60 })));
  check('expired token -> 401', r.status === 401, show(r));

  // RBAC: /api/admin/users needs "admin" or "support"
  r = await call('/api/admin/users');
  check('RBAC: valid token without the role -> 403', r.status === 403 && JSON.stringify(r.body).includes('Insufficient role'), show(r));
  check('RBAC: 403 body does not reveal the required roles', !JSON.stringify(r.body).includes('support'), show(r));
  r = await call('/api/admin/users', bearer(sign({ roles: ['support'] })));
  check('RBAC: token with one of the roles -> 200', r.status === 200, show(r));
  r = await call('/api/admin/users', { headers: { authorization: '' } });
  check('RBAC: missing token -> 401', r.status === 401, show(r));

  r = await call('/api/users/fail');
  check('backend 5xx -> 502 UPSTREAM_ERROR', r.status === 502 && r.body.code === 'UPSTREAM_ERROR', show(r));
  check('backend 5xx body not leaked', !JSON.stringify(r.body).includes('db.js'), show(r));

  r = await call('/api/products/slow');
  check('timeout -> 504 UPSTREAM_TIMEOUT', r.status === 504 && r.body.code === 'UPSTREAM_TIMEOUT', show(r));
  check('timeout bounded by timeout x attempts (2 x 2000 ms)', r.ms < 5000, `${r.ms} ms`);

  r = await call('/api/inventory');
  check('unreachable -> 503 UPSTREAM_UNAVAILABLE', r.status === 503 && r.body.code === 'UPSTREAM_UNAVAILABLE', show(r));

  // failureThreshold is 3; one failure above, two more open the breaker
  await call('/api/inventory');
  await call('/api/inventory');
  r = await call('/api/inventory');
  check('breaker open -> 503 CIRCUIT_OPEN with Retry-After', r.status === 503 && r.body.code === 'CIRCUIT_OPEN' && Number(r.headers.get('retry-after')) > 0, show(r) + ` retry-after=${r.headers.get('retry-after')}`);
  check('open breaker fails fast', r.ms < 100, `${r.ms} ms`);

  r = await call('/api/users/42');
  check('other backends unaffected by an open breaker', r.status === 200, show(r));

  console.log(failed ? `\n${failed} check(s) failed` : '\nall checks passed');
  process.exit(failed ? 1 : 0);
})();
