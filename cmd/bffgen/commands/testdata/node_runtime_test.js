// Behavioural tests for the Node.js runtime that generated routes use (upstream.js and friends).
// Run by TestNodeRuntimeJS: node node_runtime_test.js <project-dir>
const test = require('node:test');
const assert = require('node:assert');
const http = require('node:http');
const path = require('node:path');

const projectDir = process.argv[2];
const { createUpstream, aggregate, buildPath, queryStringOf } = require(path.join(projectDir, 'src/utils/upstream'));

// Minimal stand-in for an Express response / Fastify reply
function fakeReply() {
  return {
    statusCode: 200, headers: {}, body: undefined,
    status(code) { this.statusCode = code; return this; },
    header(name, value) { this.headers[name.toLowerCase()] = value; return this; },
    send(body) { this.body = body; return this; }
  };
}

function fakeRequest(overrides = {}) {
  return { method: 'GET', url: '/', params: {}, headers: {}, ...overrides };
}

// Start a backend; handler(req, res, body) answers each request
async function backend(handler) {
  const calls = [];
  const server = http.createServer((req, res) => {
    let body = '';
    req.on('data', c => { body += c; });
    req.on('end', () => { calls.push({ method: req.method, url: req.url, headers: req.headers, body }); handler(req, res, body); });
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  return { url: `http://127.0.0.1:${server.address().port}`, calls, close: () => server.close() };
}

const json = (res, status, obj, headers = {}) => {
  res.writeHead(status, { 'content-type': 'application/json', ...headers });
  res.end(JSON.stringify(obj));
};

let counter = 0;
const upstreamFor = (baseURL, options = {}) =>
  createUpstream(`svc${++counter}`, { envKey: `SVC${counter}`, baseURL, ...options });

test('aggregate combines sources and degrades', async () => {
  const b = await backend((req, res) => {
    if (req.url === '/users/42') return json(res, 200, { id: '42' });
    if (req.url === '/text') { res.writeHead(200, { 'content-type': 'text/plain' }); return res.end('hello'); }
    if (req.url === '/empty') { res.writeHead(204); return res.end(); }
    res.writeHead(200, { 'content-type': 'application/json' });
    return res.end('{not json');
  });
  try {
    const up = upstreamFor(b.url);
    const reply = fakeReply();
    await aggregate(fakeRequest({ params: { id: '42' }, headers: { 'if-none-match': '"v1"', accept: 'text/html' } }), reply, [
      { as: 'user', upstream: up, path: '/users/:id', required: true },
      { as: 'text', upstream: up, path: '/text', required: false },
      { as: 'empty', upstream: up, path: '/empty', required: false },
      { as: 'bad', upstream: up, path: '/bad', required: false }
    ]);
    assert.strictEqual(reply.statusCode, 200);
    assert.deepStrictEqual(reply.body, {
      data: { user: { id: '42' }, text: 'hello', empty: null, bad: null },
      sources: { user: 'ok', text: 'ok', empty: 'ok', bad: 'UPSTREAM_ERROR' }
    });
    assert.strictEqual(reply.headers['cache-control'], 'no-store');
    // Conditional headers apply to one resource; sources are asked for JSON
    assert.strictEqual(b.calls[0].headers['if-none-match'], undefined);
    assert.strictEqual(b.calls[0].headers.accept, 'application/json');
  } finally {
    b.close();
  }
});

test('aggregate: a failed required source decides, optional 4xx degrades', async () => {
  const b = await backend((req, res) => req.url.startsWith('/users') ? json(res, 404, { error: 'no such user' }) : json(res, 200, {}));
  try {
    const up = upstreamFor(b.url);
    let reply = fakeReply();
    await aggregate(fakeRequest({ params: { id: '7' } }), reply, [
      { as: 'orders', upstream: up, path: '/orders', required: false },
      { as: 'user', upstream: up, path: '/users/:id', required: true }
    ]);
    assert.strictEqual(reply.statusCode, 404);
    assert.strictEqual(reply.body.toString(), '{"error":"no such user"}');

    reply = fakeReply();
    await aggregate(fakeRequest({ params: { id: '7' } }), reply, [
      { as: 'orders', upstream: up, path: '/orders', required: false },
      { as: 'user', upstream: up, path: '/users/:id', required: false }
    ]);
    assert.strictEqual(reply.statusCode, 200);
    assert.deepStrictEqual(reply.body.sources, { orders: 'ok', user: 'UPSTREAM_CLIENT_ERROR' });
  } finally {
    b.close();
  }
});

test('aggregate: every source failing fails the aggregate', async () => {
  const down = upstreamFor('http://127.0.0.1:9', { retries: 0 });
  const reply = fakeReply();
  await aggregate(fakeRequest(), reply, [
    { as: 'a', upstream: down, path: '/a', required: false },
    { as: 'b', upstream: down, path: '/b', required: false }
  ]);
  assert.strictEqual(reply.statusCode, 503);
  assert.strictEqual(reply.body.code, 'UPSTREAM_UNAVAILABLE');
});

test('buildPath substitutes and encodes params', () => {
  assert.strictEqual(buildPath('/users/:id/orders/:orderId', { id: '42', orderId: '7' }), '/users/42/orders/7');
  assert.strictEqual(buildPath('/users/:id', { id: 'a/b c' }), '/users/a%2Fb%20c');
  assert.throws(() => buildPath('/users/:id', {}), /Missing path parameter "id"/);
});

test('queryStringOf keeps the raw query string', () => {
  assert.strictEqual(queryStringOf('/api/users?limit=5&tag=a&tag=b'), '?limit=5&tag=a&tag=b');
  assert.strictEqual(queryStringOf('/api/users'), '');
});

test('proxy forwards path params, query, allowlisted headers and body', async () => {
  const b = await backend((req, res) => json(res, 201, { ok: true }, { etag: '"v1"', 'x-internal': 'secret' }));
  try {
    const reply = fakeReply();
    await upstreamFor(b.url + '/api').proxy(fakeRequest({
      method: 'POST',
      url: '/api/users/42/notes?draft=true',
      params: { id: '42' },
      headers: { authorization: 'Bearer t', cookie: 'session=1', 'x-request-id': 'trace-1' },
      body: { text: 'hi' }
    }), reply, { method: 'POST', path: '/users/:id/notes' });

    const call = b.calls[0];
    assert.strictEqual(call.url, '/api/users/42/notes?draft=true');
    assert.strictEqual(call.headers.authorization, 'Bearer t');
    assert.strictEqual(call.headers['x-request-id'], 'trace-1');
    assert.strictEqual(call.headers.cookie, undefined, 'cookies must not reach backends');
    assert.strictEqual(call.body, '{"text":"hi"}');
    assert.strictEqual(call.headers['content-type'], 'application/json');

    assert.strictEqual(reply.statusCode, 201);
    assert.strictEqual(reply.headers.etag, '"v1"');
    assert.strictEqual(reply.headers['x-internal'], undefined);
    assert.deepStrictEqual(JSON.parse(reply.body), { ok: true });
  } finally { b.close(); }
});

test('proxy passes backend 4xx responses through', async () => {
  const b = await backend((req, res) => json(res, 404, { error: 'not found' }));
  try {
    const reply = fakeReply();
    await upstreamFor(b.url).proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
    assert.strictEqual(reply.statusCode, 404);
    assert.deepStrictEqual(JSON.parse(reply.body), { error: 'not found' });
  } finally { b.close(); }
});

test('GET and HEAD never forward a body', async () => {
  const b = await backend((req, res) => json(res, 200, {}));
  try {
    const up = upstreamFor(b.url);
    await up.proxy(fakeRequest({ body: { a: 1 } }), fakeReply(), { method: 'GET', path: '/x' });
    const reply = fakeReply();
    await up.proxy(fakeRequest({ method: 'HEAD' }), reply, { method: 'GET', path: '/x' });
    assert.strictEqual(b.calls[0].body, '');
    assert.strictEqual(b.calls[1].method, 'HEAD');
    assert.strictEqual(reply.body, undefined);
  } finally { b.close(); }
});

test('<ENVKEY>_URL overrides the configured base URL', async () => {
  const b = await backend((req, res) => json(res, 200, {}));
  try {
    process.env.ENVTEST_URL = b.url;
    const up = createUpstream('envtest', { envKey: 'ENVTEST', baseURL: 'http://127.0.0.1:9' });
    const reply = fakeReply();
    await up.proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
    assert.strictEqual(reply.statusCode, 200);
  } finally { b.close(); delete process.env.ENVTEST_URL; }
});

const HttpClient = require(path.join(projectDir, 'src/services/httpClient'));
const CircuitBreaker = require(path.join(projectDir, 'src/utils/circuit-breaker'));

test('timeouts are per attempt and bounded by (retries + 1) x timeout', async () => {
  const b = await backend(() => { /* never answers */ });
  try {
    const client = new HttpClient(b.url, { name: 'slow', timeout: 100, retries: 1, retryDelay: 10 });
    const started = Date.now();
    await assert.rejects(client.send('/x'), { code: 'UPSTREAM_TIMEOUT', service: 'slow' });
    const elapsed = Date.now() - started;
    assert.strictEqual(b.calls.length, 2);
    assert.ok(elapsed >= 200 && elapsed < 1000, `took ${elapsed} ms`);
  } finally { b.close(); }
});

test('connection failures are UPSTREAM_UNAVAILABLE', async () => {
  const client = new HttpClient('http://127.0.0.1:9', { name: 'down', retries: 0 });
  await assert.rejects(client.send('/x'), { code: 'UPSTREAM_UNAVAILABLE' });
});

test('5xx responses are retried for GET but never for POST', async () => {
  const b = await backend((req, res) => json(res, 503, { error: 'busy' }));
  try {
    const client = new HttpClient(b.url, { name: 'busy', retries: 2, retryDelay: 1 });
    await assert.rejects(client.send('/x'), { code: 'UPSTREAM_ERROR', upstreamStatus: 503 });
    assert.strictEqual(b.calls.length, 3);
    await assert.rejects(client.send('/x', { method: 'POST', body: '{}' }), { code: 'UPSTREAM_ERROR' });
    assert.strictEqual(b.calls.length, 4, 'POST must not be retried');
  } finally { b.close(); }
});

test('a retry that succeeds returns the response', async () => {
  let n = 0;
  const b = await backend((req, res) => (++n === 1 ? json(res, 502, {}) : json(res, 200, { ok: true })));
  try {
    const client = new HttpClient(b.url, { retries: 1, retryDelay: 1 });
    assert.deepStrictEqual(await client.get('/x'), { ok: true });
  } finally { b.close(); }
});

test('request() throws HttpError with the backend body for 4xx', async () => {
  const b = await backend((req, res) => json(res, 422, { field: 'name' }));
  try {
    const client = new HttpClient(b.url);
    await assert.rejects(client.post('/x', { a: 1 }), { name: 'HttpError', status: 422, body: { field: 'name' } });
  } finally { b.close(); }
});

test('circuit breaker opens after consecutive failures and rejects fast', async () => {
  const breaker = new CircuitBreaker({ name: 'b', failureThreshold: 2, resetTimeout: 1000 });
  const fail = () => Promise.reject(Object.assign(new Error('down'), { code: 'UPSTREAM_UNAVAILABLE' }));
  await assert.rejects(breaker.execute(fail));
  await assert.rejects(breaker.execute(fail));
  assert.strictEqual(breaker.state, 'OPEN');

  let called = false;
  await assert.rejects(breaker.execute(() => { called = true; }), err => {
    assert.strictEqual(err.code, 'CIRCUIT_OPEN');
    assert.ok(err.retryAfterMs > 0 && err.retryAfterMs <= 1000);
    return true;
  });
  assert.strictEqual(called, false);
});

test('4xx errors do not count as breaker failures', async () => {
  const breaker = new CircuitBreaker({ failureThreshold: 1 });
  await assert.rejects(breaker.execute(() => Promise.reject(Object.assign(new Error('nope'), { status: 404 }))));
  assert.strictEqual(breaker.state, 'CLOSED');
});

test('half-open admits one trial request; success closes, failure reopens', async () => {
  const breaker = new CircuitBreaker({ failureThreshold: 1, resetTimeout: 20 });
  await assert.rejects(breaker.execute(() => Promise.reject(new Error('down'))));
  await new Promise(r => setTimeout(r, 30));

  let release;
  const trial = breaker.execute(() => new Promise(r => { release = r; }));
  assert.strictEqual(breaker.state, 'HALF_OPEN');
  await assert.rejects(breaker.execute(() => 'second'), { code: 'CIRCUIT_OPEN' });
  release('ok');
  assert.strictEqual(await trial, 'ok');
  assert.strictEqual(breaker.state, 'CLOSED');

  await assert.rejects(breaker.execute(() => Promise.reject(new Error('down'))));
  await new Promise(r => setTimeout(r, 30));
  await assert.rejects(breaker.execute(() => Promise.reject(new Error('still down'))));
  assert.strictEqual(breaker.state, 'OPEN');
});

test('each backend has its own breaker, shared by its routes', async () => {
  const up = upstreamFor('http://127.0.0.1:9', { timeout: 500, retries: 0, circuitBreaker: { failureThreshold: 1, resetTimeout: 60000 } });
  const healthy = await backend((req, res) => json(res, 200, {}));
  try {
    const first = fakeReply();
    await up.proxy(fakeRequest(), first, { method: 'GET', path: '/a' });
    assert.strictEqual(first.body.code, 'UPSTREAM_UNAVAILABLE');
    const second = fakeReply();
    await up.proxy(fakeRequest(), second, { method: 'GET', path: '/b' });
    assert.strictEqual(second.body.code, 'CIRCUIT_OPEN');

    const reply = fakeReply();
    await upstreamFor(healthy.url).proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
    assert.strictEqual(reply.statusCode, 200);
  } finally { healthy.close(); }
});

test('failures are answered with 504, 503 or 502 and a consistent body', async () => {
  const slow = await backend(() => { /* never answers */ });
  const failing = await backend((req, res) => json(res, 500, { stack: 'at db.js:42' }));
  try {
    const cases = [
      [upstreamFor(slow.url, { timeout: 50, retries: 0 }), 504, 'Gateway Timeout', 'UPSTREAM_TIMEOUT'],
      [upstreamFor('http://127.0.0.1:9', { retries: 0 }), 503, 'Service Unavailable', 'UPSTREAM_UNAVAILABLE'],
      [upstreamFor(failing.url, { retries: 0 }), 502, 'Bad Gateway', 'UPSTREAM_ERROR']
    ];
    for (const [up, status, error, code] of cases) {
      const reply = fakeReply();
      await up.proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
      assert.strictEqual(reply.statusCode, status, code);
      assert.strictEqual(reply.body.error, error);
      assert.strictEqual(reply.body.code, code);
      assert.strictEqual(reply.body.service, up.name);
      assert.ok(!JSON.stringify(reply.body).includes('db.js'), 'backend 5xx body leaked');
    }
  } finally { slow.close(); failing.close(); }
});

test('an open circuit is answered with 503 and Retry-After', async () => {
  const up = upstreamFor('http://127.0.0.1:9', { retries: 0, circuitBreaker: { failureThreshold: 1, resetTimeout: 4500 } });
  await up.proxy(fakeRequest(), fakeReply(), { method: 'GET', path: '/x' });
  const reply = fakeReply();
  await up.proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
  assert.strictEqual(reply.statusCode, 503);
  assert.strictEqual(reply.body.code, 'CIRCUIT_OPEN');
  assert.strictEqual(reply.headers['retry-after'], '5');
});

test('errors carry statusCode for the framework error handlers', () => {
  assert.strictEqual(new HttpClient.UpstreamError('UPSTREAM_TIMEOUT', 'x').statusCode, 504);
  assert.strictEqual(new HttpClient.UpstreamError('UPSTREAM_UNAVAILABLE', 'x').statusCode, 503);
  assert.strictEqual(new HttpClient.UpstreamError('UPSTREAM_ERROR', 'x').statusCode, 502);
  assert.strictEqual(new CircuitBreaker.CircuitOpenError('x', 1000).statusCode, 503);
});

test('unexpected errors are not swallowed', async () => {
  const up = upstreamFor('http://127.0.0.1:9');
  await assert.rejects(up.proxy(fakeRequest(), fakeReply(), { method: 'GET', path: '/users/:id' }), /Missing path parameter/);
});

// --- authorize.js (RBAC) ---
const { authorize, userRoles, hasAnyRole } = require(path.join(projectDir, 'src/middleware/authorize'));

function withRolesClaim(claim, fn) {
  const old = process.env.JWT_ROLES_CLAIM;
  if (claim === undefined) delete process.env.JWT_ROLES_CLAIM; else process.env.JWT_ROLES_CLAIM = claim;
  try { return fn(); } finally {
    if (old === undefined) delete process.env.JWT_ROLES_CLAIM; else process.env.JWT_ROLES_CLAIM = old;
  }
}

test('userRoles reads roles, then role, by default', () => {
  withRolesClaim(undefined, () => {
    assert.deepStrictEqual(userRoles({ roles: ['admin', 'editor'] }), ['admin', 'editor']);
    assert.deepStrictEqual(userRoles({ role: 'user' }), ['user']);
    assert.deepStrictEqual(userRoles({ roles: 'admin editor,viewer' }), ['admin', 'editor', 'viewer']);
    assert.deepStrictEqual(userRoles({ roles: ['admin', 7, '', null] }), ['admin']);
    assert.deepStrictEqual(userRoles({ sub: 'u1' }), []);
    assert.deepStrictEqual(userRoles(null), []);
  });
});

test('userRoles reads JWT_ROLES_CLAIM: full name, then dotted path', () => {
  const user = {
    'cognito:groups': ['ops'],
    realm_access: { roles: ['kc-admin'] },
    'https://example.com/roles': ['auth0-admin'],
    roles: ['ignored']
  };
  assert.deepStrictEqual(withRolesClaim('cognito:groups', () => userRoles(user)), ['ops']);
  assert.deepStrictEqual(withRolesClaim('realm_access.roles', () => userRoles(user)), ['kc-admin']);
  assert.deepStrictEqual(withRolesClaim('https://example.com/roles', () => userRoles(user)), ['auth0-admin']);
  assert.deepStrictEqual(withRolesClaim('missing.claim', () => userRoles(user)), []);
});

test('hasAnyRole needs one of the roles', () => {
  assert.strictEqual(hasAnyRole({ roles: ['editor'] }, ['admin', 'editor']), true);
  assert.strictEqual(hasAnyRole({ roles: ['viewer'] }, ['admin', 'editor']), false);
});

test('authorize: next() with a role, 403 without, 401 without a user', () => {
  const check = authorize(['admin']);
  let called = 0;
  const next = () => { called++; };

  let reply = fakeReply();
  check(fakeRequest({ user: { roles: ['admin'] } }), reply, next);
  assert.strictEqual(called, 1);
  assert.strictEqual(reply.body, undefined);

  reply = fakeReply();
  check(fakeRequest({ user: { role: 'user' } }), reply, next);
  assert.strictEqual(called, 1);
  assert.strictEqual(reply.statusCode, 403);
  assert.deepStrictEqual(reply.body, { error: 'Forbidden', message: 'Insufficient role', code: 'INSUFFICIENT_ROLE' });

  reply = fakeReply();
  check(fakeRequest(), reply, next);
  assert.strictEqual(called, 1);
  assert.strictEqual(reply.statusCode, 401);
  assert.strictEqual(reply.body.code, 'UNAUTHENTICATED');
});
