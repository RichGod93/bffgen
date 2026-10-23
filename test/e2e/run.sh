#!/usr/bin/env bash
# End-to-end check of a freshly generated BFF: init + sample backends (fixtures.py) + generate,
# then lint, the project's own tests, a dependency audit, and the HTTP checks in checks.js
# against the mock backends in mock.js.
#
#   test/e2e/run.sh <nodejs-express|nodejs-fastify|python-fastapi> <workdir>
#
# BFFGEN   CLI to test (default: bffgen on PATH)
# PYTHON   interpreter for the FastAPI virtualenv (default: python3)
# AUDIT=0  skip npm audit / pip-audit (they need network access to the advisory databases)
set -uo pipefail

lang="${1:?usage: $0 <lang> <workdir>}"
work="${2:?usage: $0 <lang> <workdir>}"
here="$(cd "$(dirname "$0")" && pwd)"
bffgen="${BFFGEN:-bffgen}"
python="${PYTHON:-python3}"
port="${PORT:-8181}"
status=0

# step <name> <command...>: run quietly, print PASS/FAIL, show the output on failure
step() {
  local name="$1"; shift
  if "$@" >"$work/step.log" 2>&1; then
    echo "PASS  $name"
  else
    echo "FAIL  $name"; tail -n 40 "$work/step.log" | sed 's/^/      /'; status=1
  fi
}

# production_login_404: a production server with a real secret doesn't serve the demo login
production_login_404() {
  local p=$((port + 1)) pid code=000
  PORT=$p NODE_ENV=production JWT_SECRET=0123456789abcdef0123456789abcdef01234567 node src/index.js & pid=$!
  for _ in $(seq 100); do curl -fs "localhost:$p/health" >/dev/null && break; sleep 0.2; done
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST -H 'content-type: application/json' \
    -d '{"email":"dev@example.com","password":"secret12"}' "localhost:$p/api/auth/login")
  kill $pid 2>/dev/null; wait $pid 2>/dev/null
  echo "login status: $code"; [ "$code" = 404 ]
}

mkdir -p "$work" && work="$(cd "$work" && pwd)"
rm -rf "$work/app"
(cd "$work" && yes "" | head -n 50 | "$bffgen" init app --lang "$lang" --no-tui >"$work/init.log" 2>&1)
[ -d "$work/app" ] || { echo "FAIL  bffgen init"; cat "$work/init.log"; exit 1; }
cd "$work/app" || exit 1

case "$lang" in
  nodejs-express|nodejs-fastify)
    python3 -I "$here/fixtures.py" node bffgen.config.json
    step "bffgen generate" "$bffgen" generate
    step "bffgen generate --check (idempotent)" "$bffgen" generate --check
    [ -d node_modules ] || step "npm install" npm install --no-audit --no-fund
    step "npm run lint" npm run lint
    step "npm test" npx jest --silent
    [ "${AUDIT:-1}" = 0 ] || step "npm audit (production deps, high+)" npm audit --omit=dev --audit-level=high
    # src/middleware/auth.js runs its JWT_SECRET checks when src/index.js is loaded
    step "production: refuses to start without JWT_SECRET" sh -c 'NODE_ENV=production JWT_SECRET= node -e "require(\"./src/index\")" 2>&1 | grep "JWT_SECRET must be set"'
    step "production: refuses the .env.example placeholder secret" sh -c 'NODE_ENV=production JWT_SECRET=$(sed -n "s/^JWT_SECRET=//p" .env.example) node -e "require(\"./src/index\")" 2>&1 | grep "JWT_SECRET is a bffgen placeholder"'
    step "production: refuses a short JWT_SECRET" sh -c 'NODE_ENV=production JWT_SECRET=short node -e "require(\"./src/index\")" 2>&1 | grep "JWT_SECRET must be at least 32"'
    step "production: no demo login (404)" production_login_404
    start=(env PORT="$port" JWT_SECRET=testsecret NODE_ENV=test node src/index.js)
    export BFF_CHECKS=node
    ;;
  python-fastapi)
    python3 -I "$here/fixtures.py" python bffgen.config.py.json
    step "bffgen generate" "$bffgen" generate
    step "bffgen generate --check (idempotent)" "$bffgen" generate --check
    step "pip install" sh -c "'$python' -m venv .venv && .venv/bin/pip install -q --upgrade pip && .venv/bin/pip install -q -r requirements.txt"
    step "ruff check" .venv/bin/ruff check .
    step "pytest" .venv/bin/pytest -q -p no:cacheprovider --no-cov
    if [ "${AUDIT:-1}" != 0 ]; then
      step "pip-audit" sh -c "'$python' -m venv '$work/audit-venv' && '$work/audit-venv/bin/pip' install -q pip-audit &&
        PIPAPI_PYTHON_LOCATION=.venv/bin/python '$work/audit-venv/bin/pip-audit' --progress-spinner off"
    fi
    step "production: refuses to start without JWT_SECRET" sh -c 'ENVIRONMENT=production JWT_SECRET= .venv/bin/python -c "import config" 2>&1 | grep "JWT_SECRET must be set"'
    step "production: refuses the .env placeholder secret" sh -c 'ENVIRONMENT=production .venv/bin/python -c "import config" 2>&1 | grep "JWT_SECRET is a bffgen placeholder"'
    start=(env JWT_SECRET=testsecret .venv/bin/uvicorn main:app --port "$port" --log-level warning)
    export NO_ENCODED_SLASH=1 # Starlette decodes %2F before routing
    ;;
  *) echo "unsupported language: $lang" >&2; exit 2 ;;
esac

node "$here/mock.js" & mock=$!
"${start[@]}" >"$work/server.log" 2>&1 & server=$!
trap 'kill $mock $server 2>/dev/null' EXIT
for _ in $(seq 100); do curl -fs "localhost:$port/health" >/dev/null && break; sleep 0.2; done
node "$here/checks.js" "http://localhost:$port" || { status=1; echo "--- server log"; tail -n 40 "$work/server.log"; }
exit $status
