#!/usr/bin/env bash
#
# WeBond - turn the simulated live prices on or off (local demo).
#
#   ./scripts/live-prices.sh on      prices move every minute; gold is tradeable
#   ./scripts/live-prices.sh off     prices stop moving
#   ./scripts/live-prices.sh status
#
# Why: gold can only be traded on a price less than 5 minutes old (rule EX-10).
# The price generator is off by default, so a few minutes after start-up gold
# shows "price unavailable" until it is switched on. The setting is stored in
# the database and survives restarts. Uses the demo admin account.
#
set -euo pipefail

API="${WEBOND_API:-http://localhost:8080/api}"
ACTION="${1:-status}"

login() {
  curl -fsS -X POST "$API/auth/login" -H 'Content-Type: application/json' \
    -d '{"email":"admin@webond.test","password":"Admin@1234"}' \
    | sed -E 's/.*"accessToken":"([^"]+)".*/\1/'
}

curl -fsS -o /dev/null "${API%/api}/health/ready" 2>/dev/null \
  || { echo "ERROR: the API is not running at $API (start it with ./scripts/dev-up.sh)" >&2; exit 1; }
TOKEN="$(login)" || { echo "ERROR: could not sign in as the demo admin" >&2; exit 1; }
AUTH="Authorization: Bearer $TOKEN"

case "$ACTION" in
  on)
    curl -fsS -X POST "$API/admin/price-generator" -H "$AUTH" -H 'Content-Type: application/json' \
      -d '{"enabled":true,"intervalSeconds":60,"volatilityPct":"0.5"}' >/dev/null
    curl -fsS -X POST "$API/admin/tick-gold" -H "$AUTH" >/dev/null # fresh gold price right away
    echo "live prices ON (updates every 60 s)"
    ;;
  off)
    curl -fsS -X POST "$API/admin/price-generator" -H "$AUTH" -H 'Content-Type: application/json' \
      -d '{"enabled":false,"intervalSeconds":60,"volatilityPct":"0.5"}' >/dev/null
    echo "live prices OFF"
    ;;
  status)
    curl -fsS "$API/admin/price-generator" -H "$AUTH"
    echo
    ;;
  -h|--help) sed -n '2,13p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//' ;;
  *) echo "usage: $0 on|off|status" >&2; exit 2 ;;
esac
