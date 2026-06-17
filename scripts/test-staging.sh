#!/bin/bash
# Launch the freshly-built staging bundle in /tmp for manual testing.
# Does NOT touch ~/Applications/yiwo-draft-viewer.app — that's only updated
# by the explicit publish flow (make build + rsync).
set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STAGING_APP="$REPO_ROOT/build/staging/yiwo-draft-viewer.app"
TEST_DIR="/tmp/yiwo-test"
TEST_APP="$TEST_DIR/yiwo-draft-viewer.app"
LSREG="/System/Library/Frameworks/CoreServices.framework/Versions/A/Frameworks/LaunchServices.framework/Versions/A/Support/lsregister"

# Kill any previously-launched test instance
pkill -f "/tmp/yiwo-test/yiwo-draft-viewer.app" 2>/dev/null || true
sleep 1

# Validate staging exists
if [ ! -d "$STAGING_APP" ]; then
    echo "ERROR: $STAGING_APP does not exist."
    echo "Run 'make staging' first."
    exit 1
fi

# Stage to /tmp
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"
cp -R "$STAGING_APP" "$TEST_APP"

# Register with Launch Services so window inspection works
"$LSREG" -u -f "$TEST_APP" 2>/dev/null || true

# Launch
open "$TEST_APP"
sleep 3

# Show PID + window title
PID=$(pgrep -f "$TEST_APP" | head -1)
TITLE=$(osascript -e 'tell application "System Events" to get title of window 1 of (first process whose name contains "yiwo-draft-viewer")' 2>/dev/null || echo "(could not read title)")

echo ""
echo "✅ Test bundle launched"
echo "   Location: $TEST_APP"
echo "   PID:      ${PID:-not running}"
echo "   Title:    $TITLE"
echo ""
echo "To kill when done:"
echo "   pkill -f /tmp/yiwo-test/yiwo-draft-viewer.app"
