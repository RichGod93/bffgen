"""Write the e2e sample backends into a generated project's config.

users/orders/products point at mock.js (:4001-4003); "inventory" points at :4009, where nothing
listens (unreachable backend). Short timeouts and a low breaker threshold keep checks.js fast.

usage: python3 fixtures.py node bffgen.config.json
       python3 fixtures.py python bffgen.config.py.json
"""
import json
import sys

mode, path = sys.argv[1], sys.argv[2]


def endpoint(name, method, exposed, upstream):
    if mode == "node":
        return {"name": name, "path": upstream, "method": method, "exposeAs": exposed, "requiresAuth": True}
    return {"name": name, "method": method, "path": exposed, "upstreamPath": upstream, "requiresAuth": True}


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
    service("users", 4001, [endpoint("create_user", "POST", "/api/users", "/users")]),
    service("orders", 4002, [endpoint("get_item", "GET", "/api" + items, items)]),
    service("products", 4003),
    service("inventory", 4009),
]
with open(path, "w") as f:
    json.dump(config, f, indent=2)
