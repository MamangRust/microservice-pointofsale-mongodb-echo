#!/usr/bin/env bash
# Launch POS services locally (DB already seeded) + run hurl suite — all in
# one process tree so the services stay alive while the suite runs.
set -uo pipefail
ROOT="/home/hoover/Projects/golang/go/echo/Untitled_Folder/monolith-pointofsale-grpc"
cd "$ROOT"

LOCAL_SERVICES=(auth role user category cashier merchant order_item order product transaction apigateway)
LOGS_DIR="$ROOT/logs"
mkdir -p "$LOGS_DIR"

pkill -f 'bin/(auth|role|user|category|cashier|merchant|order_item|order|product|transaction|apigateway)' 2>/dev/null || true
sleep 1

set -a
# shellcheck disable=SC1091
source "$ROOT/.env"
set +a
export KAFKA_BROKERS=localhost:29092 APP_ENV=development

for s in "${LOCAL_SERVICES[@]}"; do
  nohup "$ROOT/bin/$s" >> "$LOGS_DIR/$s.log" 2>&1 &
done
echo "launched ${#LOCAL_SERVICES[@]} services"

GRPC_PORTS="50051 50052 50053 50054 50055 50056 50057 50058 50059 50060"
READY=0
for _ in $(seq 1 40); do
  ok=1
  for p in $GRPC_PORTS; do
    if ! timeout 1 bash -c "</dev/tcp/127.0.0.1/$p" 2>/dev/null; then ok=0; break; fi
  done
  if [ "$ok" -eq 1 ] && curl -sf "http://localhost:5000/api/auth/hello" >/dev/null 2>&1; then READY=1; break; fi
  sleep 2
done
echo "ready=$READY"
if [ "$READY" != "1" ]; then
  echo "!!! services not ready; logs with errors:"
  grep -l "panic\|fatal\|address already in use" "$LOGS_DIR"/*.log 2>/dev/null
  for f in "$LOGS_DIR"/*.log; do
    if grep -q "panic\|fatal" "$f" 2>/dev/null; then echo "-- $f"; tail -6 "$f"; fi
  done
  pkill -f 'bin/(auth|role|user|category|cashier|merchant|order_item|order|product|transaction|apigateway)' 2>/dev/null || true
  exit 1
fi

cd "$ROOT/tests/hurl"
UUID_VAL="$(uuidgen)"
PASS=0; FAIL=0; FAILED=()
for f in $(ls *.hurl | grep -v '^trace_smoke\.hurl$'); do
  echo "  → $f"
  if hurl --variable baseUrl="http://localhost:5000" --variable uuid="$UUID_VAL" --test "$f" > /tmp/pos_hurl_$f.log 2>&1; then
    PASS=$((PASS + 1)); echo "  ✅ PASS $f"
  else
    FAIL=$((FAIL + 1)); FAILED+=("$f"); echo "  ❌ FAIL $f"
    tail -8 /tmp/pos_hurl_$f.log
  fi
  sleep 2
done

echo
echo "== RESULT =="
echo "PASS: $PASS   FAIL: $FAIL   TOTAL: $((PASS + FAIL))"
if [ "$FAIL" -gt 0 ]; then
  for f in "${FAILED[@]}"; do echo "  - $f"; done
fi

pkill -f 'bin/(auth|role|user|category|cashier|merchant|order_item|order|product|transaction|apigateway)' 2>/dev/null || true
echo "cleaned up"
exit $(( FAIL > 0 ))
