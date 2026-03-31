#!/bin/bash
# build-framework.sh
# Builds the Future.xcframework from modified Xray-core source with anti-DPI protocol modifications.
#
# Prerequisites:
#   - macOS with Xcode installed
#   - Go 1.26+ (https://go.dev/dl/)
#   - gomobile: go install golang.org/x/mobile/cmd/gomobile@latest
#   - gobind:   go install golang.org/x/mobile/cmd/gobind@latest
#   - Run: gomobile init
#
# Usage:
#   ./build-framework.sh
#
# The script will:
#   1. Clone Xray-core (v26.3.27)
#   2. Apply anti-DPI protocol modifications from Xray-core-mod/
#   3. Build Future.xcframework via gomobile bind
#   4. Replace the existing framework in xFutureTunnel/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
XRAY_VERSION="v26.3.27"
XRAY_REPO="https://github.com/XTLS/Xray-core.git"
BUILD_DIR="/tmp/xray-build-$$"
OUTPUT="$SCRIPT_DIR/xFutureTunnel/Future.xcframework"

echo "=== Building Future.xcframework with anti-DPI modifications ==="
echo "Xray-core version: $XRAY_VERSION"
echo "Build directory:   $BUILD_DIR"

# Check prerequisites
for cmd in go gomobile gobind; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "ERROR: $cmd not found. Please install it first."
        exit 1
    fi
done

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
if [[ "$(printf '%s\n' "1.26" "$GO_VERSION" | sort -V | head -1)" != "1.26" ]]; then
    echo "ERROR: Go 1.26+ required, found $GO_VERSION"
    exit 1
fi

# Clone Xray-core
echo ""
echo "Step 1: Cloning Xray-core $XRAY_VERSION..."
rm -rf "$BUILD_DIR"
git clone --depth 1 --branch "$XRAY_VERSION" "$XRAY_REPO" "$BUILD_DIR"

# Apply modifications
echo ""
echo "Step 2: Applying anti-DPI protocol modifications..."

MOD_DIR="$SCRIPT_DIR/Xray-core-mod"

# Modified files
cp "$MOD_DIR/proxy/vless/encryption/common.go"   "$BUILD_DIR/proxy/vless/encryption/common.go"
cp "$MOD_DIR/proxy/vless/encryption/client.go"    "$BUILD_DIR/proxy/vless/encryption/client.go"
cp "$MOD_DIR/proxy/vless/encryption/server.go"    "$BUILD_DIR/proxy/vless/encryption/server.go"
cp "$MOD_DIR/proxy/vless/encoding/addons.go"      "$BUILD_DIR/proxy/vless/encoding/addons.go"
cp "$MOD_DIR/proxy/vless/outbound/outbound.go"    "$BUILD_DIR/proxy/vless/outbound/outbound.go"
cp "$MOD_DIR/proxy/vmess/aead/encrypt.go"         "$BUILD_DIR/proxy/vmess/aead/encrypt.go"
cp "$MOD_DIR/transport/internet/reality/reality.go" "$BUILD_DIR/transport/internet/reality/reality.go"
cp "$MOD_DIR/transport/internet/finalmask/fragment/conn.go" "$BUILD_DIR/transport/internet/finalmask/fragment/conn.go"

# New files
cp "$MOD_DIR/proxy/vless/encryption/scatter.go"   "$BUILD_DIR/proxy/vless/encryption/scatter.go"
cp "$MOD_DIR/proxy/vless/encryption/heartbeat.go"  "$BUILD_DIR/proxy/vless/encryption/heartbeat.go"

# Future gomobile binding package
mkdir -p "$BUILD_DIR/future"
cp "$MOD_DIR/future/future.go" "$BUILD_DIR/future/future.go"

echo "  Applied modifications:"
echo "    - Record size distribution (10/20/70 three-tier randomization)"
echo "    - ScatterConn TCP fragmentation (64-512 byte chunks, 0-2ms jitter)"
echo "    - HeartbeatConn idle keepalive (5-15s intervals)"
echo "    - VLESS header padding (16-64 bytes random)"
echo "    - VMess AEAD padding (0-32 bytes trailing)"
echo "    - SessionID timestamp jitter (±300s for REALITY)"
echo "    - Post-handshake packet fragmentation (packets 2-4)"
echo "    - HeartbeatConn integration for non-XTLS flows"

# Verify build
echo ""
echo "Step 3: Verifying Go build..."
cd "$BUILD_DIR"
go build ./...
echo "  Go build: OK"

# Build xcframework
echo ""
echo "Step 4: Building Future.xcframework..."
gomobile bind -v \
    -target=ios,macos \
    -o "$BUILD_DIR/Future.xcframework" \
    ./future/

echo "  Framework built: $BUILD_DIR/Future.xcframework"

# Replace existing framework
echo ""
echo "Step 5: Replacing existing framework..."
rm -rf "$OUTPUT"
cp -R "$BUILD_DIR/Future.xcframework" "$OUTPUT"

echo ""
echo "=== Build complete ==="
echo "Framework: $OUTPUT"
echo ""
echo "To verify, open xNetFuture.xcworkspace in Xcode and build."

# Cleanup
rm -rf "$BUILD_DIR"
echo "Build directory cleaned up."
