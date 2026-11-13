#!/usr/bin/env bash
#
# WeBond — start the whole stack (MySQL + API + web) with one command.
#
#   ./scripts/dev-up.sh              build and start everything
#   ./scripts/dev-up.sh --no-build   start without rebuilding the images
#
# The API creates the database tables and loads the demo data itself on first
# boot, so nothing else needs to be run. Stop with: docker compose stop
#
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

BUILD="--build"
for arg in "$@"; do
  case "$arg" in
    --no-build) BUILD="" ;;
    -h|--help) sed -n '2,10p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown option: $arg (try --help)" >&2; exit 2 ;;
  esac
done

API_URL="http://localhost:8080"
WEB_URL="http://localhost:5173"

step() { printf '\n==> %s\n' "$*"; }
fail() { printf '\nERROR: %s\n' "$*" >&2; exit 1; }

# --- 1. Docker must be installed and running --------------------------------
step "Checking Docker"
command -v docker >/dev/null 2>&1 || fail "Docker is not installed. Install Docker Desktop: https://www.docker.com/products/docker-desktop/"
if ! docker info >/dev/null 2>&1; then
  fail "The Docker daemon is not running. Open Docker Desktop, wait until it says 'running', then re-run this script."
fi
docker compose version >/dev/null 2>&1 || fail "'docker compose' (v2) is not available. Update Docker Desktop."
echo "Docker OK"

# --- 2. .env with random secrets --------------------------------------------
step "Preparing .env"
./scripts/gen-env.sh

# --- 3. Start ---------------------------------------------------------------
step "Starting MySQL, API and web (first run builds the images; this can take a few minutes)"
# shellcheck disable=SC2086  # $BUILD is intentionally unquoted: empty or --build
docker compose up -d $BUILD

# --- 4. Wait until it answers -----------------------------------------------
wait_for() {
  local name="$1" url="$2" tries="${3:-90}" i
  printf 'Waiting for %s ' "$name"
  for ((i = 1; i <= tries; i++)); do
    if curl -fsS -o /dev/null "$url" 2>/dev/null; then
      echo "ready"
      return 0
    fi
    printf '.'
    sleep 2
  done
  echo
  return 1
}

step "Waiting for the services"
if ! wait_for "API" "$API_URL/health/ready"; then
  docker compose logs --tail=40 api || true
  fail "The API did not become ready. See the log above, or run: docker compose logs api"
fi
if ! wait_for "web" "$WEB_URL/" 30; then
  docker compose logs --tail=20 web || true
  fail "The web UI did not start. Run: docker compose logs web"
fi

# --- 5. Done ----------------------------------------------------------------
cat <<DONE

WeBond is running.

  Web UI : $WEB_URL
  API    : $API_URL   (health: $API_URL/health/ready)
  E-mail : http://localhost:8025   (local inbox for verification / reset e-mails)

  Demo investor : demo@webond.test  /  Demo@1234
  Demo admin    : admin@webond.test /  Admin@1234   (API only, no admin screens)

  Stop    : docker compose stop
  Restart : ./scripts/dev-up.sh --no-build
  Logs    : docker compose logs -f api
  Reset   : docker compose down -v   (DELETES the database)
DONE
