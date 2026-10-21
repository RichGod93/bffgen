// Mock downstreams users/orders/products on :4001-4003 (MOCK_PORTS=4001,4002,4003 to change).
// Echoes method/url/body as JSON. Special paths (anywhere in the URL):
//   slow    -> responds after 10 s        fail  -> 500 with a "stack trace" body
//   missing -> 404 JSON                   empty -> 204
const http = require('http');
const ports = (process.env.MOCK_PORTS || '4001,4002,4003').split(',').map(Number);
const names = ['users', 'orders', 'products'];
ports.forEach((port, i) => {
  const name = names[i] || `svc${i}`;
  http.createServer((req, res) => {
    let body = ''; req.on('data', c => body += c); req.on('end', () => {
      const send = (status, obj) => {
        res.writeHead(status, { 'content-type': 'application/json', 'x-internal': 'secret', 'cache-control': 'max-age=5' });
        res.end(JSON.stringify(obj));
      };
      if (req.url.includes('slow')) return setTimeout(() => send(200, { service: name, slow: true }), 10000);
      if (req.url.includes('fail')) return send(500, { error: 'boom', stack: 'at db.query (/srv/app/db.js:42)' });
      if (req.url.includes('missing')) return send(404, { error: 'not found', service: name });
      if (req.url.includes('empty')) { res.writeHead(204); return res.end(); }
      send(200, { service: name, method: req.method, url: req.url, body, auth: !!req.headers.authorization, requestId: req.headers['x-request-id'] || null });
    });
  }).listen(port);
});
