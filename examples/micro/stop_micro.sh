#!/usr/bin/env bash
set -euo pipefail

# stop_micro.sh
# Stops micro/master/cluster demo processes and removes pid files.

LOG_DIR="/tmp"

echo "Stopping micro demo processes (by pid files and process name)..."

for f in master game gate; do
  pidfile="$LOG_DIR/nano_${f}.pid"
  if [ -f "$pidfile" ]; then
    pid=$(cat "$pidfile")
    echo " KILL $f pid $pid"
    # try graceful, then force
    kill "$pid" 2>/dev/null || true
    sleep 0.1
    if kill -0 "$pid" 2>/dev/null; then
      echo " PID $pid still exists, forcing kill -9"
      kill -9 "$pid" 2>/dev/null || true
    fi
    rm -f "$pidfile"
  fi
done

# Fallbacks: kill by known process patterns. go run may spawn temporary
# binaries under /tmp/go-build*/b*/exe/main, so include patterns to catch them.
pkill -f "examples/micro/main.go" || true
pkill -f "examples/cluster/main.go" || true
pkill -f "/tmp/go-build.*/b.*/exe/main" || true
pkill -f "exe/main gate" || true
pkill -f "exe/main game" || true
pkill -f "exe/main master" || true

echo "Stopped. Logs are kept in /tmp/nano_*.log"
