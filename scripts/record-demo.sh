#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
TMP_DIR=/tmp/oc-demo-recording

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

TOOLS_DIR="$TMP_DIR/tools"
VENV_DIR="$TOOLS_DIR/venv"
WORK_DIR="$TMP_DIR/work"
BIN_DIR="$TMP_DIR/bin"
HOME_DIR="$TMP_DIR/home"
CONFIG_DIR="$TMP_DIR/config"
CAST_PATH="$REPO_ROOT/docs/demo.cast"
SVG_PATH="$REPO_ROOT/docs/demo.svg"
SESSION_SCRIPT="$TMP_DIR/demo-session.sh"

rm -rf "$TMP_DIR"
mkdir -p "$TOOLS_DIR" "$WORK_DIR" "$BIN_DIR" "$HOME_DIR" "$CONFIG_DIR"

python3 -m venv "$VENV_DIR"
"$VENV_DIR/bin/pip" install --quiet asciinema

go build -o "$BIN_DIR/oc" "$REPO_ROOT"

cat > "$SESSION_SCRIPT" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

cd "$OC_DEMO_WORKDIR"

run_cmd() {
  printf '$ %s\n' "$*"
  sleep 0.6
  /bin/zsh -c "$*"
  sleep 0.9
}

mkdir -p demo-project

run_cmd 'oc --version'
run_cmd 'oc source add qbicsoftware/opencode-config-bundle --name qbic'
run_cmd 'oc source list'
run_cmd 'oc bundle apply qbic --preset mixed --project-root demo-project'
run_cmd 'oc bundle status --project-root demo-project'
run_cmd 'ls demo-project'
EOF

chmod +x "$SESSION_SCRIPT"

env \
  HOME="$HOME_DIR" \
  OC_DEMO_WORKDIR="$WORK_DIR" \
  XDG_CONFIG_HOME="$CONFIG_DIR" \
  PATH="$BIN_DIR:$PATH" \
  TERM="xterm-256color" \
  "$VENV_DIR/bin/asciinema" rec \
  --overwrite \
  --quiet \
  --idle-time-limit 1.2 \
  --cols 100 \
  --rows 26 \
  -c "$SESSION_SCRIPT" \
  "$CAST_PATH"

npx -p svg-term-cli svg-term \
  --in "$CAST_PATH" \
  --out "$SVG_PATH" \
  --width 100 \
  --height 26 \
  --padding 12 \
  --window

printf 'written: %s\n' "$CAST_PATH"
printf 'written: %s\n' "$SVG_PATH"
