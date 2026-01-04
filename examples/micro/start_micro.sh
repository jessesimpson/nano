#!/usr/bin/env bash
set -euo pipefail

# start_micro.sh
# Stops any old demo processes, then starts micro master/game/gate in background.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="/tmp"

echo "Stopping any existing cluster/micro demo processes (via stop_micro.sh)..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STOP_SH="$SCRIPT_DIR/stop_micro.sh"
if [ -x "$STOP_SH" ]; then
  "$STOP_SH"
else
  # fallback to execute with bash in case not executable
  bash "$STOP_SH" || true
fi
sleep 0.5

echo "Starting micro master..."
cd "$ROOT_DIR"
nohup go run ./examples/micro/main.go master > "$LOG_DIR/nano_master.log" 2>&1 &
echo $! > "$LOG_DIR/nano_master.pid"
sleep 0.3

echo "Starting micro game..."
nohup go run ./examples/micro/main.go game --listen 127.0.0.1:34568 --master 127.0.0.1:34567 > "$LOG_DIR/nano_game.log" 2>&1 &
echo $! > "$LOG_DIR/nano_game.pid"
sleep 0.3

echo "Starting micro gate..."
nohup go run ./examples/micro/main.go gate --listen 127.0.0.1:34569 --gate-address 127.0.0.1:34590 --master 127.0.0.1:34567 > "$LOG_DIR/nano_gate.log" 2>&1 &
echo $! > "$LOG_DIR/nano_gate.pid"

sleep 1

echo "Started services. PIDs:"
for f in master game gate; do
  pidfile="$LOG_DIR/nano_${f}.pid"
  if [ -f "$pidfile" ]; then
    printf "  %-6s %s\n" "$f" "$(cat $pidfile)"
  fi
done

echo "Recent logs (tail 20):"
tail -n 20 "$LOG_DIR/nano_master.log" || true
tail -n 20 "$LOG_DIR/nano_game.log" || true
tail -n 20 "$LOG_DIR/nano_gate.log" || true

echo "You can run the test client with: go run ./wsclient/main.go"
