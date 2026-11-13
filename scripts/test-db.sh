#!/usr/bin/env bash
#
# WeBond - create (or recreate) the separate database used by the API tests.
#
#   ./scripts/test-db.sh            create webond_test if missing and migrate it
#   ./scripts/test-db.sh --reset    drop and recreate it (wipes test data only)
#
# The API tests create users, orders and DCA rounds and run system-wide jobs, so
# they must never run against the application database "webond". This script
# prepares "webond_test" on the docker compose MySQL and prints the variable to
# export. The demo data in "webond" is not touched.
#
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

TEST_DB="webond_test"
RESET=0
for arg in "$@"; do
  case "$arg" in
    --reset) RESET=1 ;;
    -h|--help) sed -n '2,12p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown option: $arg (try --help)" >&2; exit 2 ;;
  esac
done

# Read a value from .env (falls back to the docker-compose default).
env_value() {
  local v=""
  [ -f .env ] && v="$(grep -E "^$1=" .env | head -n1 | cut -d= -f2- || true)"
  echo "${v:-$2}"
}
DB_USER="$(env_value MYSQL_USER webond)"
DB_PASS="$(env_value MYSQL_PASSWORD webond)"
ROOT_PASS="$(env_value MYSQL_ROOT_PASSWORD webond-root)"

docker info >/dev/null 2>&1 || { echo "ERROR: Docker is not running" >&2; exit 1; }

echo "==> Starting MySQL"
docker compose up -d mysql >/dev/null
for _ in $(seq 1 60); do
  if docker compose exec -T mysql mysqladmin ping -uroot -p"$ROOT_PASS" --silent >/dev/null 2>&1; then break; fi
  sleep 2
done

sql() { docker compose exec -T mysql mysql -uroot -p"$ROOT_PASS" -e "$1" 2>/dev/null; }

if [ "$RESET" -eq 1 ]; then
  echo "==> Dropping $TEST_DB"
  sql "DROP DATABASE IF EXISTS \`$TEST_DB\`;"
fi
echo "==> Creating $TEST_DB (if missing) and granting $DB_USER"
sql "CREATE DATABASE IF NOT EXISTS \`$TEST_DB\` CHARACTER SET utf8mb4;
     GRANT ALL PRIVILEGES ON \`$TEST_DB\`.* TO '$DB_USER'@'%';
     FLUSH PRIVILEGES;"

echo "==> Applying migrations to $TEST_DB (using the API image)"
docker compose build api >/dev/null
docker compose run --rm --no-deps \
  -e DATABASE_URL="$DB_USER:$DB_PASS@tcp(mysql:3306)/$TEST_DB?parseTime=true&loc=Asia%2FBangkok" \
  api migrate

cat <<DONE

Test database ready. Run the API tests with:

  export WEBOND_TEST_DSN='$DB_USER:$DB_PASS@tcp(127.0.0.1:3306)/$TEST_DB?parseTime=true&loc=Asia%2FBangkok'
  export WEBOND_TEST_ROOT_DSN='root:$ROOT_PASS@tcp(127.0.0.1:3306)/'   # optional: migration tests
  cd api && go test -tags=integration ./...
DONE
