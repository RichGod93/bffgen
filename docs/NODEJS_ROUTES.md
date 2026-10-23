# Generated Node.js Routes

> Python (FastAPI) projects generate routes with the same behaviour; see
> [Python/FastAPI Support](PYTHON_SUPPORT.md#what-generated-routes-do).

`bffgen generate` turns each backend in `bffgen.config.json` into a route file,
`src/routes/<backend>.js`, for Express or Fastify, and the `aggregates` into
`src/routes/aggregates.js`. This page describes what those routes do: how requests are
forwarded, how timeouts, retries and the circuit breaker work, how aggregates combine several
backends, and which status codes clients see when a backend fails.

## Configuration

The config format is described by the JSON Schema
[`schemas/bffgen-config.json`](../schemas/bffgen-config.json), which new configs reference in
`$schema` so editors can validate and autocomplete them. `bffgen config validate` checks a config
with the same rules as `bffgen generate`.

```json
{
  "backends": [
    {
      "name": "users",
      "baseUrl": "http://localhost:4001/api",
      "timeout": 5000,
      "retries": 2,
      "circuitBreaker": { "enabled": true, "failureThreshold": 5, "resetTimeout": 30000 },
      "endpoints": [
        { "name": "getUser", "method": "GET", "path": "/users/:id", "exposeAs": "/api/users/:id", "requiresAuth": true }
      ]
    }
  ]
}
```

| Field | Meaning | Default |
|---|---|---|
| `timeout` | Per-attempt timeout in ms, or a duration string such as `"5s"` | `10000` |
| `retries` | Extra attempts for `GET`, `HEAD` and `OPTIONS` requests | `2` |
| `circuitBreaker.enabled` | Use a circuit breaker for this backend | `true` |
| `circuitBreaker.failureThreshold` | Consecutive failures that open the circuit | `5` |
| `circuitBreaker.resetTimeout` | Time in ms the circuit stays open before a trial request | `30000` |

The values in effect are written into the generated route file. Endpoint paths may use
`:id` or `{id}` parameters. Every parameter in `path` must also appear in `exposeAs`;
`generate` reports an error otherwise.

### Environment overrides

Each backend's settings can be overridden at runtime. The prefix is the backend name in upper
case with other characters replaced by `_` (`user-service` → `USER_SERVICE`).

| Variable | Overrides |
|---|---|
| `USERS_URL` (or `USERS_SERVICE_URL`) | `baseUrl` |
| `USERS_TIMEOUT` | `timeout` (ms) |
| `USERS_RETRIES` | `retries` |
| `USERS_CIRCUIT_FAILURE_THRESHOLD` | `circuitBreaker.failureThreshold` |
| `USERS_CIRCUIT_RESET_TIMEOUT` | `circuitBreaker.resetTimeout` (ms) |

## What gets forwarded

For `GET /api/users/42?fields=name&tag=a&tag=b` with the endpoint above, the backend receives
`GET http://localhost:4001/api/users/42?fields=name&tag=a&tag=b`.

- **Path parameters** are substituted and URL-encoded.
- **The query string** is forwarded as received, so repeated keys and their order are kept.
- **Request headers:** `authorization`, `accept`, `accept-language`, `if-none-match`,
  `if-modified-since` and `x-request-id` (the caller's, or the framework's request id).
  Cookies and other headers are not forwarded.
- **The request body** is forwarded for every method except `GET` and `HEAD`.
- **Responses below 500** are passed back unchanged: status, body, and the `content-type`,
  `cache-control`, `etag`, `last-modified` and `location` headers. A backend `404` or `400`
  reaches the client as is.

## Timeouts and retries

Each attempt has its own timeout, which also covers reading the response body. Only
idempotent methods (`GET`, `HEAD`, `OPTIONS`) are retried, after a timeout, a connection
failure or a 5xx response, with exponential backoff starting at 100 ms. A `POST`, `PUT`,
`PATCH` or `DELETE` is sent once. The longest a `GET` can take is about
`(retries + 1) × timeout`.

## Circuit breaker

Each backend has one circuit breaker, shared by all of its routes, so one failing backend
doesn't affect the others.

- **Closed:** requests pass through. After `failureThreshold` consecutive failures (timeouts,
  connection failures or 5xx responses, counted after retries) the circuit opens. 4xx
  responses don't count as failures.
- **Open:** requests are rejected immediately with `503` and a `Retry-After` header, without
  calling the backend.
- **Half-open:** after `resetTimeout`, one trial request is let through. If it succeeds the
  circuit closes; if it fails the circuit opens again.

## Authentication and roles (RBAC)

An endpoint with `"requiresAuth": true` needs a valid JWT. An endpoint with `"roles"` also
needs one of the listed roles; `roles` implies `requiresAuth`:

```json
{ "name": "deleteUser", "method": "DELETE", "path": "/users/:id", "exposeAs": "/api/users/:id", "roles": ["admin", "support"] }
```

| Request | Status | Body |
|---|---|---|
| No token | `401 Unauthorized` | `{ "error": "Unauthorized", "message": "Authentication required", "code": "UNAUTHENTICATED" }` |
| Expired token | `401 Unauthorized` | `{ ..., "code": "TOKEN_EXPIRED" }` |
| Bad signature, other algorithm, malformed, or `nbf` in the future | `401 Unauthorized` | `{ ..., "code": "INVALID_TOKEN" }` |
| Valid token without any of the roles | `403 Forbidden` | `{ "error": "Forbidden", "message": "Insufficient role", "code": "INSUFFICIENT_ROLE" }` |
| Valid token with one of the roles | proxied to the backend | |

### Tokens and `JWT_SECRET`

One module, `src/middleware/auth.js`, checks tokens for generated routes, aggregates and
`/api/auth/profile`. Use its `authenticate` in your own routes too (Express middleware, or a
Fastify `onRequest` hook; Fastify projects also have it as `fastify.authenticate`).

- The token is read from `Authorization: Bearer <token>` (the scheme is case-insensitive), else
  from the httpOnly `access_token` cookie that the login route sets.
- Only HS256 tokens signed with `JWT_SECRET` are accepted (`algorithms: ['HS256']`); `exp` and `nbf`
  are checked. The verified payload is `req.user` (Fastify: `request.user`).
- With `NODE_ENV=production` the server refuses to start unless `JWT_SECRET` is set to at least
  32 characters (RFC 7518 §3.2) that are not one of bffgen's placeholders. Generate one with
  `openssl rand -base64 48`.
- Outside production an unset `JWT_SECRET` becomes a random secret for that process, with a
  warning; tokens then stop working when the server restarts.
- `POST /api/auth/login` is a demo that signs a token for any email and password. It is not
  served when `NODE_ENV=production`; `AUTH_DEMO_LOGIN=true` or `false` overrides that.
  `signToken(payload)` from `auth.js` signs tokens the same way for your own login.

### Roles

The 403 response doesn't say which roles are required. Routes call `authorize([...])` from
`src/middleware/authorize.js` after authentication (Express: route middleware, Fastify:
`preHandler`); you can use it in your own routes too.

Roles are read from the verified token:

- By default from the `roles` claim (a list, as in RFC 9068), or from `role` if there is no `roles` claim.
- With `JWT_ROLES_CLAIM` set, from that claim instead, e.g. `cognito:groups` (Amazon Cognito) or
  a dotted path such as `realm_access.roles` (Keycloak). A claim whose name contains dots, such as an
  Auth0 namespaced claim `https://example.com/roles`, is matched by its full name first.
- A claim can be a list of strings or one string of space- or comma-separated roles, so role names
  can't contain spaces or commas.

The demo `/api/auth/login` issues tokens with `role: "user"`. Replace it with your identity
provider, or issue tokens with the roles your endpoints need.

Projects created before v3.10.0 have a framework-specific `src/middleware/auth.js` without a
`bffgen:runtime` marker; `bffgen generate` replaces it (previous copy: `auth.js.bak`), and
regenerated Fastify routes use it instead of the check in `src/index.js`. `src/index.js` is
yours, so its own login, profile check and fallback secret stay until you update them (the
3.10.0 CHANGELOG lists the steps); `generate` warns while they are there.

Projects created before v3.8.0 answered an invalid or expired token with 403 in Fastify routes
(the check in `src/index.js`). Once an endpoint has `roles`, `bffgen generate` changes that line
to 401, so 403 always means a missing role; nothing else in `src/index.js` is changed, and the
previous copy is saved as `src/index.js.bak`. If `src/middleware/authorize.js` is a file of your
own (no `bffgen:runtime` marker), it is kept, and `generate` asks you to rename it before it
generates endpoints with `roles`.

401 responses carry `WWW-Authenticate: Bearer` (RFC 9110), with `error="invalid_token"` for an
invalid, expired or not-yet-valid token.

## Aggregates

An aggregate is a `GET` route that calls several backends in parallel and answers with their
combined responses, so a frontend makes one request instead of several. Aggregates are listed at
the top level of the config and refer to backends by name:

```json
{
  "backends": [
    { "name": "users", "baseUrl": "http://localhost:4001/api", "endpoints": [] },
    { "name": "orders", "baseUrl": "http://localhost:4002/api", "timeout": 2000 },
    { "name": "products", "baseUrl": "http://localhost:4003/api" }
  ],
  "aggregates": [
    {
      "name": "dashboard",
      "exposeAs": "/api/dashboard/:id",
      "requiresAuth": true,
      "sources": [
        { "backend": "users", "path": "/users/:id", "as": "user", "required": true },
        { "backend": "orders", "path": "/orders?userId=:id", "as": "orders" },
        { "backend": "products", "path": "/products/recommended", "as": "recommendations" }
      ]
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `exposeAs` | The path the BFF exposes (`GET`, and `HEAD`). Parameters as `:id` or `{id}`. |
| `name` | Optional; used in comments and (Python) handler names |
| `requiresAuth`, `roles` | As for endpoints (401 / 403, see above) |
| `sources[].backend` | A backend from `backends`. Its `timeout`, `retries`, circuit breaker and environment overrides apply. A backend used only by aggregates needs no `endpoints`. |
| `sources[].path` | `GET` path on that backend. It may carry a query string; its parameters come from `exposeAs`. |
| `sources[].as` | Key of this source in the response (default: the backend name) |
| `sources[].required` | If `true`, a failure of this source fails the whole request (default `false`) |

A successful request answers `200` with each source's parsed body (JSON, or text for other
content types; `null` for an empty body) and its status:

```json
{
  "data": { "user": { "id": "42" }, "orders": [], "recommendations": null },
  "sources": { "user": "ok", "orders": "ok", "recommendations": "UPSTREAM_TIMEOUT" }
}
```

| Situation | Response |
|---|---|
| Every source succeeds | `200`, every source `"ok"` |
| An optional source fails | `200` with that source's data `null` and its status set to the error code: `UPSTREAM_TIMEOUT`, `UPSTREAM_UNAVAILABLE`, `CIRCUIT_OPEN`, `UPSTREAM_ERROR` (5xx or invalid JSON) or `UPSTREAM_CLIENT_ERROR` (4xx). Sent with `Cache-Control: no-store`, so a partial response isn't cached as a complete one. |
| A required source fails | The aggregate fails as a proxied route would: `502`/`503`/`504` with the error body below, or the source's 4xx response passed through (a missing user gives `404`) |
| Every source fails | As if the sources were required: there is nothing to return |

When several sources decide the response, the first in config order wins, so the result is
deterministic. A request takes as long as its slowest source, which is bounded by that backend's
timeout and retries. Sources get the same request headers as proxied routes (`Authorization`,
`Accept-Language`, `X-Request-Id`, …), except the conditional headers (`If-None-Match`,
`If-Modified-Since`), which apply to a single resource, and with `Accept: application/json`. The
client's query string is not forwarded, because it isn't clear which source it is meant for; put
parameters in `exposeAs` and use them in the source paths instead.

`generate` rejects aggregates that refer to an unknown backend, use a parameter `exposeAs` doesn't
define, use the same `as` twice, or have the same path as another aggregate or a `GET` endpoint.
With aggregates configured, no backend may be named `aggregates`. `bffgen generate-docs`
documents aggregates in the OpenAPI spec. Python projects generate the same routes in
`routers/aggregates_router.py`.

For aggregation that needs your own logic (transforming or joining the data, calls that depend on
each other), write a controller; see the [Node.js Aggregation Guide](NODEJS_AGGREGATION.md).

## Error responses

When a backend fails, the route answers with one of these statuses and a JSON body:

| Status | `code` | When |
|---|---|---|
| `504 Gateway Timeout` | `UPSTREAM_TIMEOUT` | No response within the timeout, on every attempt |
| `503 Service Unavailable` | `UPSTREAM_UNAVAILABLE` | The backend could not be reached (connection refused, DNS failure, reset) |
| `503 Service Unavailable` | `CIRCUIT_OPEN` | The circuit is open; `Retry-After` gives the seconds until the next trial |
| `502 Bad Gateway` | `UPSTREAM_ERROR` | The backend answered with a 5xx status |

```json
{ "error": "Gateway Timeout", "message": "users did not respond within 5000 ms", "code": "UPSTREAM_TIMEOUT", "service": "users" }
```

The body of a backend 5xx response is not passed on, because it may contain internal details
such as stack traces.

## Runtime files

Generated routes use these files in your project:

| File | Purpose |
|---|---|
| `src/utils/upstream.js` | Forwards requests, runs aggregates and maps failures to responses |
| `src/services/httpClient.js` | HTTP client with timeouts and retries |
| `src/utils/circuit-breaker.js` | Circuit breaker |
| `src/middleware/auth.js` | JWT authentication (since v3.10.0) |
| `src/middleware/authorize.js` | Role checks for endpoints with `roles` |

You can use `HttpClient` and `CircuitBreaker` in your own services and controllers too. Errors
they throw (`UpstreamError`, `CircuitOpenError`) carry a `statusCode` of 502, 503 or 504, so
passing them to the framework's error handler gives the same status codes as generated routes.

Each file carries a `bffgen:runtime <n>` marker. `bffgen generate` creates any that are
missing. If a file has an older marker (or none, as in projects created before v3.4.0), it is
replaced, and your previous copy is kept next to it as `<file>.bak` (`<file>.bak.1`, `.bak.2`, …
if earlier backups exist; none is overwritten). Files with the current
marker are never touched, so local edits survive until bffgen ships a new runtime version.
`--dry-run` and `--check` only report what would change.

## Regenerating

`bffgen generate` owns the files it generates for each backend: `src/routes/<backend>.js`,
`src/controllers/<backend>.controller.js`, `src/controllers/<backend>.controller.basic.js` and
`src/services/<backend>.service.js`, plus `src/routes/aggregates.js` (Python:
`routers/<backend>_router.py`, `services/<backend>_service.py`, `routers/aggregates_router.py`
and `routers/__init__.py`). You can edit any of them:

- A file that still has the content bffgen last wrote is regenerated with your config changes.
- A file you edited is **kept** and reported (`✋ Kept …`). It no longer receives config changes.
  `bffgen generate --force` regenerates it and saves your version as `<file>.bak` (or
  `<file>.bak.1`, `.bak.2`, … so an earlier backup is never overwritten). Emptying a file counts
  as an edit.
- A file for a backend you removed from the config is reported, never deleted.
- An invalid config writes nothing, and neither does any other error found before writing (for
  example a corrupt `.bffgen/generated.json`). Backend names must be unique; names that would
  give the same file on a case-insensitive disk (`users`, `Users`) or the same env var
  (`user-service`, `user_service`) are rejected too.
- Line endings don't count as edits, so a Windows checkout with `core.autocrlf=true` regenerates
  normally.

bffgen records what it wrote (a SHA-256 per file) in `.bffgen/generated.json`. Commit that file;
it contains no timestamps, so the same config always produces the same output. Projects created
before v3.7.0 don't have it yet: on the first run, files identical to the new output are recorded,
and files that differ are kept in case you edited them (run `--force` once if you didn't).
Python projects created before v3.7.0 ignored all of `.bffgen/` in `.gitignore`; `bffgen generate`
changes that line so the manifest is committed.

The route block between `// bffgen:begin:routes` and `// bffgen:end:routes` in `src/index.js` is
always rewritten; the rest of `src/index.js` is yours.

| Command | Writes | Exit code |
|---|---|---|
| `bffgen generate` | regenerates unedited files | 0 |
| `bffgen generate --dry-run` | nothing; lists what would change | 0 |
| `bffgen generate --check` | nothing; lists what would change | 1 if anything would change, or an edited file no longer matches the config (for CI) |
| `bffgen generate --force` | also regenerates edited files (`.bak` copies) | 0 |

`--check` passes on a file you edited as long as the config still generates what that file was
made from: your edits sit on top of the current output. Once the config changes what bffgen would
generate for it, `--check` fails until you regenerate it (`--force`) and copy your changes back.

