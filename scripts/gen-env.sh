#!/usr/bin/env bash
#
# WeBond — create or harden the root .env.
#
#   ./scripts/gen-env.sh            create .env from .env.example with random JWT
#                                   secrets (does nothing if .env already exists)
#   ./scripts/gen-env.sh --rotate   keep the existing .env but replace both JWT
#                                   secrets with fresh random values (this signs
#                                   everyone out)
#
# .env is gitignored. Idempotent and safe to re-run.
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT/.env"
EXAMPLE="$ROOT/.env.example"
ROTATE=0

for arg in "$@"; do
  case "$arg" in
    --rotate) ROTATE=1 ;;
    -h|--help) sed -n '2,13p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown option: $arg (try --help)" >&2; exit 2 ;;
  esac
done

# 64 hex chars = 256 bits of randomness.
random_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
  else
    LC_ALL=C tr -dc 'a-f0-9' </dev/urandom | head -c 64
  fi
}

# Replace KEY=... in $ENV_FILE without sed -i (its syntax differs on macOS/Linux).
set_var() {
  local key="$1" value="$2" tmp
  tmp="$(mktemp)"
  awk -v key="$key" -v value="$value" '
    BEGIN { done = 0 }
    $0 ~ "^" key "=" { print key "=" value; done = 1; next }
    { print }
    END { if (!done) print key "=" value }
  ' "$ENV_FILE" >"$tmp"
  cat "$tmp" >"$ENV_FILE"
  rm -f "$tmp"
}

get_var() {
  grep -E "^$1=" "$ENV_FILE" 2>/dev/null | head -n1 | cut -d= -f2- || true
}

if [ ! -f "$ENV_FILE" ]; then
  [ -f "$EXAMPLE" ] || { echo "ERROR: $EXAMPLE not found" >&2; exit 1; }
  cp "$EXAMPLE" "$ENV_FILE"
  chmod 600 "$ENV_FILE" 2>/dev/null || true
  set_var JWT_ACCESS_SECRET "$(random_secret)"
  set_var JWT_REFRESH_SECRET "$(random_secret)"
  echo "created .env with random JWT secrets"
  exit 0
fi

if [ "$ROTATE" -eq 1 ]; then
  set_var JWT_ACCESS_SECRET "$(random_secret)"
  set_var JWT_REFRESH_SECRET "$(random_secret)"
  echo "rotated JWT secrets in .env (existing sessions are now invalid)"
  exit 0
fi

echo ".env already exists; left unchanged"
for key in JWT_ACCESS_SECRET JWT_REFRESH_SECRET; do
  value="$(get_var "$key")"
  if [ -z "$value" ] || [[ "$value" == *change-me* ]]; then
    echo "WARNING: $key is empty or still the 'change-me' placeholder." >&2
    echo "         Run: ./scripts/gen-env.sh --rotate" >&2
  fi
done
