// Behavioural tests for the Node.js runtime that generated routes use (upstream.js and friends).
// Run by TestNodeRuntimeJS: node node_runtime_test.js <project-dir>
const test = require('node:test');
const assert = require('node:assert');
const http = require('node:http');
const path = require('node:path');

const projectDir = process.argv[2];
const { createUpstream, buildPath, queryStringOf } = require(path.join(projectDir, 'src/utils/upstream'));

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
    await assert.rejects(up.proxy(fakeRequest(), fakeReply(), { method: 'GET', path: '/a' }), { code: 'UPSTREAM_UNAVAILABLE' });
    await assert.rejects(up.proxy(fakeRequest(), fakeReply(), { method: 'GET', path: '/b' }), { code: 'CIRCUIT_OPEN' });

    const reply = fakeReply();
    await upstreamFor(healthy.url).proxy(fakeRequest(), reply, { method: 'GET', path: '/x' });
    assert.strictEqual(reply.statusCode, 200);
  } finally { healthy.close(); }
});
