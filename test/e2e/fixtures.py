"""Write the e2e sample backends into a generated project's config.

users/orders/products point at mock.js (:4001-4003); "inventory" points at :4009, where nothing
listens (unreachable backend). /api/admin/users requires the "admin" role (RBAC). Short timeouts and a low breaker threshold keep checks.js fast.
The aggregates combine users/orders/products (all up), add a slow and an unreachable optional
source (partial degradation), and fail every source (/api/broken).

usage: python3 fixtures.py node bffgen.config.json
       python3 fixtures.py python bffgen.config.py.json
"""
import json
import sys

mode, path = sys.argv[1], sys.argv[2]


def endpoint(name, method, exposed, upstream, roles=None):
    if mode == "node":
        ep = {"name": name, "path": upstream, "method": method, "exposeAs": exposed, "requiresAuth": True}
    else:
        ep = {"name": name, "method": method, "path": exposed, "upstreamPath": upstream, "requiresAuth": True}
    if roles:
        ep["roles"] = roles
    return ep


def service(name, port, extra=()):
    param = ":id" if mode == "node" else "{id}"
    return {
        "name": name,
        "baseUrl": f"http://localhost:{port}/api",
        "timeout": 2000,
        "retries": 1,
        "circuitBreaker": {"failureThreshold": 3, "resetTimeout": 5000},
        "endpoints": [
            endpoint(f"list_{name}", "GET", f"/api/{name}", f"/{name}"),
            endpoint(f"get_{name}", "GET", f"/api/{name}/{param}", f"/{name}/{param}"),
            *extra,
        ],
    }


with open(path) as f:
    config = json.load(f)
items = ("/orders/{orderId}/items/{itemId}" if mode == "node" else "/orders/:orderId/items/:itemId")
config["backends"] = [
    service("users", 4001, [
        endpoint("create_user", "POST", "/api/users", "/users"),
        endpoint("admin_list_users", "GET", "/api/admin/users", "/users", roles=["admin", "support"]),
    ]),
    service("orders", 4002, [endpoint("get_item", "GET", "/api" + items, items)]),
    service("products", 4003),
    service("inventory", 4009),
]


def source(backend, upstream, key, required=False):
    src = {"backend": backend, "path": upstream, "as": key}
    if required:
        src["required"] = True
    return src


p = ":id" if mode == "node" else "{id}"
config["aggregates"] = [
    {"name": "dashboard", "exposeAs": f"/api/dashboard/{p}", "requiresAuth": True, "sources": [
        source("users", f"/users/{p}", "user", required=True),
        source("orders", f"/orders?userId={p}", "orders"),
        source("products", "/products", "products"),
    ]},
    {"name": "degraded", "exposeAs": f"/api/degraded/{p}", "requiresAuth": True, "sources": [
        source("users", f"/users/{p}", "user", required=True),
        source("products", "/products/slow", "slow"),
        source("inventory", "/inventory", "inventory"),
    ]},
    {"exposeAs": "/api/broken", "sources": [
        source("inventory", "/inventory", "first"),
        source("inventory", "/inventory/other", "second"),
    ]},
    {"name": "admin-dashboard", "exposeAs": "/api/admin/dashboard", "roles": ["admin"], "sources": [
        source("users", "/users", "users"),
    ]},
]
with open(path, "w") as f:
    json.dump(config, f, indent=2)
