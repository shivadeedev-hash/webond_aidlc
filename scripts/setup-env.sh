#!/usr/bin/env bash
#
# WeBond — development environment setup / check (macOS + Homebrew).
#
# Development uses `npm run dev` (frontend mock mode, Node only) and needs
# NO Go / Docker / MySQL. Those tools are only required for the REAL/backend
# build (Go API + MySQL via Docker Compose). This script detects what is
# present, installs the missing real-build tools via Homebrew when possible,
# and prints a summary. It is idempotent and safe to re-run.
#
set -uo pipefail

# --- helpers -----------------------------------------------------------------
BOLD=$'\033[1m'; DIM=$'\033[2m'; RESET=$'\033[0m'
OK="OK"; MISSING="MISSING"; INSTALLED="INSTALLED"; SKIPPED="SKIPPED"

have() { command -v "$1" >/dev/null 2>&1; }

# status table rows collected as "tool|status|version"
ROWS=()
add_row() { ROWS+=("$1|$2|$3"); }

ver() {
  # best-effort single-line version for a tool
  case "$1" in
    node)   node -v 2>/dev/null ;;
    npm)    npm -v 2>/dev/null ;;
    git)    git --version 2>/dev/null | awk '{print $3}' ;;
    go)     go version 2>/dev/null | awk '{print $3}' ;;
    docker) docker --version 2>/dev/null | awk -F'[ ,]' '{print $3}' ;;
    "docker compose") docker compose version --short 2>/dev/null ;;
    mysql)  mysql --version 2>/dev/null | awk '{print $3}' | tr -d ',' ;;
    *) echo "" ;;
  esac
}

echo "${BOLD}WeBond environment setup${RESET}"
echo "${DIM}macOS / Homebrew · idempotent · safe to re-run${RESET}"
echo

# --- OS / Homebrew detection -------------------------------------------------
OS="$(uname -s)"
if [ "$OS" != "Darwin" ]; then
  echo "WARN: detected non-macOS ($OS). This script targets macOS + Homebrew."
  echo "      It will still CHECK tools but will not attempt installs."
fi

BREW_OK=0
if have brew; then
  BREW_OK=1
else
  echo "WARN: Homebrew (brew) not found."
  echo "      Install it from https://brew.sh then re-run this script."
  echo "      Missing tools will be reported but NOT installed automatically."
fi
echo

# --- always-required (do not auto-install these core dev tools) --------------
# Node/npm/git are expected to already be present for development.
for t in node npm git; do
  if have "$t"; then
    echo "  [check] $t ................. ${OK} ($(ver "$t"))"
    add_row "$t" "$OK" "$(ver "$t")"
  else
    echo "  [check] $t ................. ${MISSING}"
    echo "          -> please install $t manually (core dev dependency)."
    add_row "$t" "$MISSING" "-"
  fi
done
echo

# --- real-build tools (auto-install via brew when missing) -------------------
# install_tool <display> <probe-cmd> <brew-install-args...>
install_tool() {
  local display="$1"; shift
  local probe="$1"; shift
  if have "$probe"; then
    echo "  [check] $display ........ ${OK} ($(ver "$display"))"
    add_row "$display" "$OK" "$(ver "$display")"
    return
  fi

  echo "  [check] $display ........ ${MISSING}"
  if [ "$OS" != "Darwin" ] || [ "$BREW_OK" -ne 1 ]; then
    echo "          -> cannot auto-install (need macOS + Homebrew). Install manually: brew install $*"
    add_row "$display" "$MISSING" "-"
    return
  fi

  echo "          -> installing: brew install $*"
  if brew install "$@"; then
    if have "$probe"; then
      echo "          -> ${INSTALLED} ($(ver "$display"))"
      add_row "$display" "$INSTALLED" "$(ver "$display")"
    else
      # e.g. Docker Desktop cask installs an app that still needs launching
      echo "          -> installed, but '$probe' not yet on PATH (may need to launch the app)."
      add_row "$display" "$INSTALLED" "pending launch"
    fi
  else
    echo "          -> brew install failed. Install $display manually."
    add_row "$display" "$MISSING" "-"
  fi
}

# Go — required for the backend build.
install_tool "go" "go" go

# Docker Desktop — provides docker + compose v2 + a MySQL container.
if have docker; then
  echo "  [check] docker .......... ${OK} ($(ver docker))"
  add_row "docker" "$OK" "$(ver docker)"
else
  echo "  [check] docker .......... ${MISSING}"
  if [ "$OS" = "Darwin" ] && [ "$BREW_OK" -eq 1 ]; then
    echo "          -> installing: brew install --cask docker (Docker Desktop)"
    if brew install --cask docker; then
      echo "          -> ${INSTALLED}. Launch Docker Desktop once to finish setup."
      add_row "docker" "$INSTALLED" "pending launch"
    else
      echo "          -> brew install failed. Install Docker Desktop manually: https://www.docker.com/products/docker-desktop/"
      add_row "docker" "$MISSING" "-"
    fi
  else
    echo "          -> install Docker Desktop manually: https://www.docker.com/products/docker-desktop/"
    add_row "docker" "$MISSING" "-"
  fi
fi

# docker compose (v2 subcommand) — bundled with Docker Desktop.
if have docker && docker compose version >/dev/null 2>&1; then
  echo "  [check] docker compose .. ${OK} ($(ver "docker compose"))"
  add_row "docker compose" "$OK" "$(ver "docker compose")"
else
  echo "  [check] docker compose .. ${MISSING} (ships with Docker Desktop)"
  add_row "docker compose" "$MISSING" "-"
fi

# mysql client — OPTIONAL (Docker provides the server). Install if brew is available.
if have mysql; then
  echo "  [check] mysql client .... ${OK} ($(ver mysql))"
  add_row "mysql client" "$OK" "$(ver mysql)"
else
  echo "  [check] mysql client .... ${MISSING} (optional)"
  if [ "$OS" = "Darwin" ] && [ "$BREW_OK" -eq 1 ]; then
    echo "          -> installing optional client: brew install mysql-client"
    if brew install mysql-client; then
      echo "          -> ${INSTALLED} (you may need to add it to PATH per brew's caveats)."
      add_row "mysql client" "$INSTALLED" "$(ver mysql)"
    else
      echo "          -> optional install failed; skipping (server is provided by Docker)."
      add_row "mysql client" "$SKIPPED" "-"
    fi
  else
    echo "          -> optional; skipping (MySQL server is provided by Docker)."
    add_row "mysql client" "$SKIPPED" "-"
  fi
fi

# --- summary table -----------------------------------------------------------
echo
echo "${BOLD}Summary${RESET}"
printf '  %-16s | %-10s | %s\n' "tool" "status" "version"
printf '  %-16s-+-%-10s-+-%s\n' "----------------" "----------" "----------------"
for row in "${ROWS[@]}"; do
  IFS='|' read -r tool status version <<< "$row"
  printf '  %-16s | %-10s | %s\n' "$tool" "$status" "${version:--}"
done

echo
echo "Next steps:"
echo "  • Frontend demo (Node only):  cd web && npm install && npm run dev"
echo "  • Full stack (needs Docker):  cp .env.example .env && docker compose up --build"
echo "  • If Docker was just installed, launch Docker Desktop once, then re-run this script."
