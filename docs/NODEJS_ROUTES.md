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

Generated routes use three files in your project:

| File | Purpose |
|---|---|
| `src/utils/upstream.js` | Forwards requests and maps failures to responses |
| `src/services/httpClient.js` | HTTP client with timeouts and retries |
| `src/utils/circuit-breaker.js` | Circuit breaker |

You can use `HttpClient` and `CircuitBreaker` in your own services and controllers too. Errors
they throw (`UpstreamError`, `CircuitOpenError`) carry a `statusCode` of 502, 503 or 504, so
passing them to the framework's error handler gives the same status codes as generated routes.

Each file carries a `bffgen:runtime <n>` marker. `bffgen generate` creates any that are
missing. If a file has an older marker (or none, as in projects created before v3.4.0), it is
replaced, and your previous copy is kept next to it as `<file>.bak`. Files with the current
marker are never touched, so local edits survive until bffgen ships a new runtime version.
`--dry-run` and `--check` only report what would change.

Route files themselves are regenerated on every `bffgen generate`; don't edit them by hand.
