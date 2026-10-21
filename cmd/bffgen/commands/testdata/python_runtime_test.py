"""
Behavioural tests for the Python runtime that generated FastAPI routes use
(utils/upstream.py, services/http_client.py, utils/circuit_breaker.py, utils/authorize.py).
Run by TestPythonRuntime: python python_runtime_test.py <project-dir>
"""

import asyncio
import json
import os
import sys
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

sys.path.insert(0, sys.argv.pop(1))

import httpx  # noqa: E402
from fastapi import FastAPI, Request  # noqa: E402

from services.http_client import HttpClient, HttpError, UpstreamError  # noqa: E402
from utils.circuit_breaker import CircuitBreaker, CircuitOpenError  # noqa: E402
from utils.upstream import Source, aggregate, build_path, create_upstream  # noqa: E402
from fastapi import Depends  # noqa: E402
from fastapi.testclient import TestClient  # noqa: E402
from middleware.auth import require_authentication  # noqa: E402
from utils.authorize import has_any_role, require_roles, user_roles  # noqa: E402


class Backend:
    """A local HTTP backend; handler(request_info) returns (status, headers, body) or None to hang"""

    def __init__(self, handler):
        self.calls = []
        outer = self

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def handle_one(self):
                length = int(self.headers.get("content-length") or 0)
                body = self.rfile.read(length).decode() if length else ""
                call = {"method": self.command, "url": self.path, "headers": dict(self.headers.items()), "body": body}
                call["headers"] = {k.lower(): v for k, v in call["headers"].items()}
                outer.calls.append(call)
                result = handler(call)
                if result is None:
                    time.sleep(5)
                    return
                status, headers, payload = result
                self.send_response(status)
                for name, value in headers.items():
                    self.send_header(name, value)
                if payload is None:
                    data = b""
                elif isinstance(payload, bytes):
                    data = payload
                else:
                    data = json.dumps(payload).encode()
                self.send_header("content-length", str(len(data)))
                self.end_headers()
                if self.command != "HEAD":
                    self.wfile.write(data)

            do_GET = do_POST = do_PUT = do_DELETE = do_HEAD = handle_one

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
        self.server.daemon_threads = True
        self.url = f"http://127.0.0.1:{self.server.server_address[1]}"
        threading.Thread(target=self.server.serve_forever, kwargs={"poll_interval": 0.01}, daemon=True).start()

    def close(self):
        self.server.shutdown()
        self.server.server_close()


def json_response(status, payload, headers=None):
    return status, {"content-type": "application/json", **(headers or {})}, payload


counter = 0


def upstream_for(base_url, **options):
    global counter
    counter += 1
    return create_upstream(f"svc{counter}", env_key=f"SVC{counter}", base_url=base_url, **options)


def app_for(up, method, path, route="/proxy/{id}"):
    """A FastAPI app with one route that proxies like a generated router"""
    app = FastAPI()

    @app.api_route(route, methods=["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"])
    async def handler(request: Request):
        return await up.proxy(request, method, path)

    return httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://bff")


class BuildPathTest(unittest.TestCase):
    def test_substitutes_and_encodes(self):
        self.assertEqual(build_path("/users/{id}/orders/{order_id}", {"id": "42", "order_id": "7"}), "/users/42/orders/7")
        self.assertEqual(build_path("/users/{id}", {"id": "a/b c?"}), "/users/a%2Fb%20c%3F")
        with self.assertRaisesRegex(ValueError, 'Missing path parameter "id"'):
            build_path("/users/{id}", {})


class ProxyTest(unittest.IsolatedAsyncioTestCase):
    async def test_forwards_params_query_headers_and_body(self):
        backend = Backend(lambda c: json_response(201, {"ok": True}, {"etag": '"v1"', "x-internal": "secret"}))
        try:
            async with app_for(upstream_for(backend.url + "/api"), "POST", "/users/{id}/notes") as bff:
                r = await bff.post(
                    "/proxy/42?draft=true&tag=a&tag=b",
                    json={"text": "hi"},
                    headers={"authorization": "Bearer t", "cookie": "session=1", "x-request-id": "trace-1"},
                )
            call = backend.calls[0]
            self.assertEqual(call["url"], "/api/users/42/notes?draft=true&tag=a&tag=b")
            self.assertEqual(call["headers"]["authorization"], "Bearer t")
            self.assertEqual(call["headers"]["x-request-id"], "trace-1")
            self.assertNotIn("cookie", call["headers"], "cookies must not reach backends")
            self.assertEqual(json.loads(call["body"]), {"text": "hi"})
            self.assertEqual(call["headers"]["content-type"], "application/json")
            self.assertEqual(r.status_code, 201)
            self.assertEqual(r.headers["etag"], '"v1"')
            self.assertNotIn("x-internal", r.headers)
            self.assertEqual(r.json(), {"ok": True})
        finally:
            backend.close()

    async def test_passes_4xx_through(self):
        backend = Backend(lambda c: json_response(404, {"error": "not found"}))
        try:
            async with app_for(upstream_for(backend.url), "GET", "/x/{id}") as bff:
                r = await bff.get("/proxy/1")
            self.assertEqual(r.status_code, 404)
            self.assertEqual(r.json(), {"error": "not found"})
        finally:
            backend.close()

    async def test_get_and_head_never_forward_a_body(self):
        backend = Backend(lambda c: json_response(200, {}))
        try:
            async with app_for(upstream_for(backend.url), "GET", "/x/{id}") as bff:
                await bff.request("GET", "/proxy/1", content=b'{"a": 1}')
                r = await bff.head("/proxy/1")
            self.assertEqual(backend.calls[0]["body"], "")
            self.assertEqual(backend.calls[1]["method"], "HEAD")
            self.assertEqual(r.content, b"")
        finally:
            backend.close()

    async def test_env_url_overrides_base_url(self):
        backend = Backend(lambda c: json_response(200, {}))
        os.environ["ENVTEST_URL"] = backend.url
        try:
            up = create_upstream("envtest", env_key="ENVTEST", base_url="http://127.0.0.1:9")
            async with app_for(up, "GET", "/x/{id}") as bff:
                r = await bff.get("/proxy/1")
            self.assertEqual(r.status_code, 200)
        finally:
            del os.environ["ENVTEST_URL"]
            backend.close()

    async def test_failures_map_to_504_503_502(self):
        slow = Backend(lambda c: None)
        failing = Backend(lambda c: json_response(500, {"stack": "at db.py:42"}))
        try:
            cases = [
                (upstream_for(slow.url, timeout=100, retries=0), 504, "Gateway Timeout", "UPSTREAM_TIMEOUT"),
                (upstream_for("http://127.0.0.1:9", retries=0), 503, "Service Unavailable", "UPSTREAM_UNAVAILABLE"),
                (upstream_for(failing.url, retries=0), 502, "Bad Gateway", "UPSTREAM_ERROR"),
            ]
            for up, status, error, code in cases:
                async with app_for(up, "GET", "/x/{id}") as bff:
                    r = await bff.get("/proxy/1")
                self.assertEqual(r.status_code, status, code)
                body = r.json()
                self.assertEqual((body["error"], body["code"], body["service"]), (error, code, up.name))
                self.assertNotIn("db.py", r.text, "backend 5xx body leaked")
        finally:
            slow.close()
            failing.close()

    async def test_open_circuit_answers_503_with_retry_after(self):
        up = upstream_for("http://127.0.0.1:9", retries=0, circuit_breaker={"failureThreshold": 1, "resetTimeout": 4500})
        async with app_for(up, "GET", "/x/{id}") as bff:
            await bff.get("/proxy/1")
            r = await bff.get("/proxy/1")
        self.assertEqual(r.status_code, 503)
        self.assertEqual(r.json()["code"], "CIRCUIT_OPEN")
        self.assertEqual(r.headers["retry-after"], "5")

    async def test_each_backend_has_its_own_breaker(self):
        down = upstream_for("http://127.0.0.1:9", retries=0, circuit_breaker={"failureThreshold": 1, "resetTimeout": 60000})
        healthy = Backend(lambda c: json_response(200, {}))
        try:
            async with app_for(down, "GET", "/x/{id}") as bff:
                self.assertEqual((await bff.get("/proxy/1")).json()["code"], "UPSTREAM_UNAVAILABLE")
                self.assertEqual((await bff.get("/proxy/1")).json()["code"], "CIRCUIT_OPEN")
            async with app_for(upstream_for(healthy.url), "GET", "/x/{id}") as bff:
                self.assertEqual((await bff.get("/proxy/1")).status_code, 200)
        finally:
            healthy.close()


def aggregate_app(sources, route="/agg/{id}"):
    """A FastAPI app with one aggregate route, like a generated aggregates router"""
    app = FastAPI()

    @app.api_route(route, methods=["GET", "HEAD"])
    async def handler(request: Request):
        return await aggregate(request, sources)

    return httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://bff")


class AggregateTest(unittest.IsolatedAsyncioTestCase):
    async def test_combines_sources_and_degrades(self):
        backend = Backend(lambda c: {
            "/users/42": json_response(200, {"id": "42"}),
            "/text": (200, {"content-type": "text/plain"}, b"hello"),
            "/empty": (204, {}, None),
            "/bad": (200, {"content-type": "application/json"}, b"{not json"),
        }[c["url"]])
        try:
            up = upstream_for(backend.url)
            sources = [
                Source("user", up, "/users/{id}", required=True),
                Source("text", up, "/text"),
                Source("empty", up, "/empty"),
                Source("bad", up, "/bad"),
            ]
            async with aggregate_app(sources) as bff:
                r = await bff.get("/agg/42", headers={"if-none-match": '"v1"', "accept": "text/html"})
            self.assertEqual(r.status_code, 200)
            self.assertEqual(r.json(), {
                "data": {"user": {"id": "42"}, "text": "hello", "empty": None, "bad": None},
                "sources": {"user": "ok", "text": "ok", "empty": "ok", "bad": "UPSTREAM_ERROR"},
            })
            self.assertEqual(r.headers["cache-control"], "no-store")
            # Conditional headers apply to one resource; sources are asked for JSON
            self.assertNotIn("if-none-match", backend.calls[0]["headers"])
            self.assertEqual(backend.calls[0]["headers"]["accept"], "application/json")
        finally:
            backend.close()

    async def test_required_source_decides(self):
        backend = Backend(lambda c: json_response(404, {"error": "no such user"}) if "users" in c["url"] else json_response(200, {}))
        try:
            up = upstream_for(backend.url)
            async with aggregate_app([Source("orders", up, "/orders"), Source("user", up, "/users/{id}", required=True)]) as bff:
                r = await bff.get("/agg/7")
            self.assertEqual(r.status_code, 404)
            self.assertEqual(r.json(), {"error": "no such user"})

            # Optional 4xx: partial response
            async with aggregate_app([Source("orders", up, "/orders"), Source("user", up, "/users/{id}")]) as bff:
                r = await bff.get("/agg/7")
            self.assertEqual(r.status_code, 200)
            self.assertEqual(r.json()["sources"], {"orders": "ok", "user": "UPSTREAM_CLIENT_ERROR"})
        finally:
            backend.close()

    async def test_every_source_failing_fails_the_aggregate(self):
        down = upstream_for("http://127.0.0.1:9", retries=0)
        async with aggregate_app([Source("a", down, "/a"), Source("b", down, "/b")], route="/agg") as bff:
            r = await bff.get("/agg")
        self.assertEqual(r.status_code, 503)
        self.assertEqual(r.json()["code"], "UPSTREAM_UNAVAILABLE")


class HttpClientTest(unittest.IsolatedAsyncioTestCase):
    async def test_timeout_is_per_attempt_and_bounded(self):
        backend = Backend(lambda c: None)
        try:
            client = HttpClient(backend.url, name="slow", timeout=100, retries=1, retry_delay=10)
            started = time.monotonic()
            with self.assertRaises(UpstreamError) as ctx:
                await client.send("/x")
            elapsed = time.monotonic() - started
            self.assertEqual(ctx.exception.code, "UPSTREAM_TIMEOUT")
            self.assertEqual(len(backend.calls), 2)
            self.assertTrue(0.2 <= elapsed < 1.0, f"took {elapsed:.2f} s")
            await client.aclose()
        finally:
            backend.close()

    async def test_connection_failure_is_unavailable(self):
        client = HttpClient("http://127.0.0.1:9", retries=0)
        with self.assertRaises(UpstreamError) as ctx:
            await client.send("/x")
        self.assertEqual(ctx.exception.code, "UPSTREAM_UNAVAILABLE")

    async def test_5xx_retried_for_get_never_for_post(self):
        backend = Backend(lambda c: json_response(503, {}))
        try:
            client = HttpClient(backend.url, retries=2, retry_delay=1)
            with self.assertRaises(UpstreamError) as ctx:
                await client.send("/x")
            self.assertEqual((ctx.exception.code, ctx.exception.upstream_status), ("UPSTREAM_ERROR", 503))
            self.assertEqual(len(backend.calls), 3)
            with self.assertRaises(UpstreamError):
                await client.send("/x", method="POST", content=b"{}")
            self.assertEqual(len(backend.calls), 4, "POST must not be retried")
            await client.aclose()
        finally:
            backend.close()

    async def test_request_raises_http_error_for_4xx(self):
        backend = Backend(lambda c: json_response(422, {"field": "name"}))
        try:
            client = HttpClient(backend.url)
            with self.assertRaises(HttpError) as ctx:
                await client.post("/x", {"a": 1})
            self.assertEqual((ctx.exception.status, ctx.exception.body), (422, {"field": "name"}))
            await client.aclose()
        finally:
            backend.close()

    def test_errors_carry_status_code(self):
        self.assertEqual(UpstreamError("UPSTREAM_TIMEOUT", "x", "s").status_code, 504)
        self.assertEqual(UpstreamError("UPSTREAM_UNAVAILABLE", "x", "s").status_code, 503)
        self.assertEqual(UpstreamError("UPSTREAM_ERROR", "x", "s").status_code, 502)
        self.assertEqual(CircuitOpenError("s", 1).status_code, 503)


class CircuitBreakerTest(unittest.IsolatedAsyncioTestCase):
    async def fail(self):
        raise UpstreamError("UPSTREAM_UNAVAILABLE", "down", "b")

    async def test_opens_after_consecutive_failures_and_rejects_fast(self):
        breaker = CircuitBreaker(failure_threshold=2, timeout=1, name="b")
        for _ in range(2):
            with self.assertRaises(UpstreamError):
                await breaker.call(self.fail)
        called = False

        async def probe():
            nonlocal called
            called = True

        with self.assertRaises(CircuitOpenError) as ctx:
            await breaker.call(probe)
        self.assertTrue(0 < ctx.exception.retry_after <= 1)
        self.assertFalse(called)

    async def test_4xx_does_not_count(self):
        breaker = CircuitBreaker(failure_threshold=1)

        async def not_found():
            raise HttpError(404, None, "b")

        with self.assertRaises(HttpError):
            await breaker.call(not_found)
        self.assertEqual(breaker.state.value, "closed")

    async def test_half_open_admits_one_trial(self):
        breaker = CircuitBreaker(failure_threshold=1, timeout=0.02)
        with self.assertRaises(UpstreamError):
            await breaker.call(self.fail)
        await asyncio.sleep(0.03)

        release = asyncio.Event()

        async def slow_ok():
            await release.wait()
            return "ok"

        trial = asyncio.create_task(breaker.call(slow_ok))
        await asyncio.sleep(0)
        self.assertEqual(breaker.state.value, "half_open")
        with self.assertRaises(CircuitOpenError):
            await breaker.call(slow_ok)
        release.set()
        self.assertEqual(await trial, "ok")
        self.assertEqual(breaker.state.value, "closed")

        with self.assertRaises(UpstreamError):
            await breaker.call(self.fail)
        await asyncio.sleep(0.03)
        with self.assertRaises(UpstreamError):
            await breaker.call(self.fail)
        self.assertEqual(breaker.state.value, "open")


class AuthorizeTest(unittest.TestCase):
    def setUp(self):
        self.old = os.environ.pop("JWT_ROLES_CLAIM", None)

    def tearDown(self):
        os.environ.pop("JWT_ROLES_CLAIM", None)
        if self.old is not None:
            os.environ["JWT_ROLES_CLAIM"] = self.old

    def test_default_claims(self):
        self.assertEqual(user_roles({"roles": ["admin", "editor"]}), ["admin", "editor"])
        self.assertEqual(user_roles({"role": "user"}), ["user"])
        self.assertEqual(user_roles({"roles": "admin editor,viewer"}), ["admin", "editor", "viewer"])
        self.assertEqual(user_roles({"roles": ["admin", 7, "", None]}), ["admin"])
        self.assertEqual(user_roles({"sub": "u1"}), [])
        self.assertEqual(user_roles(None), [])

    def test_custom_claim(self):
        user = {
            "cognito:groups": ["ops"],
            "realm_access": {"roles": ["kc-admin"]},
            "https://example.com/roles": ["auth0-admin"],
            "roles": ["ignored"],
        }
        for claim, want in [
            ("cognito:groups", ["ops"]),
            ("realm_access.roles", ["kc-admin"]),
            ("https://example.com/roles", ["auth0-admin"]),
            ("missing.claim", []),
        ]:
            os.environ["JWT_ROLES_CLAIM"] = claim
            self.assertEqual(user_roles(user), want, claim)

    def test_has_any_role(self):
        self.assertTrue(has_any_role({"roles": ["editor"]}, ("admin", "editor")))
        self.assertFalse(has_any_role({"roles": ["viewer"]}, ("admin", "editor")))

    def test_require_roles(self):
        app = FastAPI()

        @app.get("/admin", dependencies=[Depends(require_roles("admin"))])
        async def admin():
            return {"ok": True}

        client = TestClient(app)
        app.dependency_overrides[require_authentication] = lambda: {"roles": ["admin"]}
        self.assertEqual(client.get("/admin").status_code, 200)
        app.dependency_overrides[require_authentication] = lambda: {"role": "user"}
        r = client.get("/admin")
        self.assertEqual(r.status_code, 403)
        self.assertEqual(r.json(), {"detail": "Insufficient role"})


if __name__ == "__main__":
    unittest.main(verbosity=1)
