# Generated Node.js Routes

> Python (FastAPI) projects generate routes with the same behaviour; see
> [Python/FastAPI Support](PYTHON_SUPPORT.md#what-generated-routes-do).

`bffgen generate` turns each backend in `bffgen.config.json` into a route file,
`src/routes/<backend>.js`, for Express or Fastify. This page describes what those routes do:
how requests are forwarded, how timeouts, retries and the circuit breaker work, and which
status codes clients see when a backend fails.

## Configuration

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

An endpoint with `"requiresAuth": true` needs a valid JWT (`Authorization: Bearer <token>`,
signed with `JWT_SECRET`). An endpoint with `"roles"` also needs one of the listed roles;
`roles` implies `requiresAuth`:

```json
{ "name": "deleteUser", "method": "DELETE", "path": "/users/:id", "exposeAs": "/api/users/:id", "roles": ["admin", "support"] }
```

| Request | Status | Body |
|---|---|---|
| No token, or an invalid or expired one | `401 Unauthorized` | `{ "error": "Invalid or expired token" }` (or `Authentication required`) |
| Valid token without any of the roles | `403 Forbidden` | `{ "error": "Forbidden", "message": "Insufficient role", "code": "INSUFFICIENT_ROLE" }` |
| Valid token with one of the roles | proxied to the backend | |

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

Projects created before v3.8.0 answered an invalid or expired token with 403 in Fastify routes
(the check in `src/index.js`). `bffgen generate` changes that line to 401, so 403 always means a
missing role; nothing else in `src/index.js` is changed.

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
| `src/utils/upstream.js` | Forwards requests and maps failures to responses |
| `src/services/httpClient.js` | HTTP client with timeouts and retries |
| `src/utils/circuit-breaker.js` | Circuit breaker |
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
`src/services/<backend>.service.js` (Python: `routers/<backend>_router.py`,
`services/<backend>_service.py` and `routers/__init__.py`). You can edit any of them:

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

