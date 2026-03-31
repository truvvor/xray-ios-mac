#!/bin/bash
# build-framework.sh
# Builds the Future.xcframework from libXray + modified Xray-core with anti-DPI protocol modifications.
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

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
XRAY_VERSION="v26.3.27"
XRAY_REPO="https://github.com/XTLS/Xray-core.git"
BUILD_DIR="/tmp/xray-build-$$"
OUTPUT="$SCRIPT_DIR/xFutureTunnel/Future.xcframework"

echo "=== Building Future.xcframework (libXray + anti-DPI mods) ==="

# Check prerequisites
for cmd in go gomobile gobind; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "ERROR: $cmd not found. Please install it first."
        exit 1
    fi
done

# Setup build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Step 1: Clone Xray-core and apply anti-DPI modifications
echo ""
echo "Step 1: Setting up modified Xray-core..."
git clone --depth 1 --branch "$XRAY_VERSION" "$XRAY_REPO" "$BUILD_DIR/Xray-core"

MOD_DIR="$SCRIPT_DIR/Xray-core-mod"

# Apply modified files
cp "$MOD_DIR/proxy/vless/encryption/common.go"   "$BUILD_DIR/Xray-core/proxy/vless/encryption/common.go"
cp "$MOD_DIR/proxy/vless/encryption/client.go"    "$BUILD_DIR/Xray-core/proxy/vless/encryption/client.go"
cp "$MOD_DIR/proxy/vless/encryption/server.go"    "$BUILD_DIR/Xray-core/proxy/vless/encryption/server.go"
cp "$MOD_DIR/proxy/vless/encoding/addons.go"      "$BUILD_DIR/Xray-core/proxy/vless/encoding/addons.go"
cp "$MOD_DIR/proxy/vless/outbound/outbound.go"    "$BUILD_DIR/Xray-core/proxy/vless/outbound/outbound.go"
cp "$MOD_DIR/proxy/vmess/aead/encrypt.go"         "$BUILD_DIR/Xray-core/proxy/vmess/aead/encrypt.go"
cp "$MOD_DIR/transport/internet/reality/reality.go" "$BUILD_DIR/Xray-core/transport/internet/reality/reality.go"
cp "$MOD_DIR/transport/internet/finalmask/fragment/conn.go" "$BUILD_DIR/Xray-core/transport/internet/finalmask/fragment/conn.go"

# Apply new files
cp "$MOD_DIR/proxy/vless/encryption/scatter.go"   "$BUILD_DIR/Xray-core/proxy/vless/encryption/scatter.go"
cp "$MOD_DIR/proxy/vless/encryption/heartbeat.go"  "$BUILD_DIR/Xray-core/proxy/vless/encryption/heartbeat.go"

echo "  Anti-DPI modifications applied."

# Step 2: Copy libXray with replace directive pointing to local Xray-core
echo ""
echo "Step 2: Setting up libXray..."
cp -r "$SCRIPT_DIR/libXray" "$BUILD_DIR/libXray"

# Update replace directive to point to build dir
cd "$BUILD_DIR/libXray"
sed -i.bak "s|replace github.com/xtls/xray-core => ../Xray-core|replace github.com/xtls/xray-core => $BUILD_DIR/Xray-core|" go.mod
rm -f go.mod.bak

echo "  libXray configured."

# Step 3: Verify build
echo ""
echo "Step 3: Verifying Go build..."
cd "$BUILD_DIR/libXray"
go build $(go list ./... | grep -v build/template | grep -v download_geo | grep -v desktop_bin)
echo "  Go build: OK"

# Step 4: Build xcframework via gomobile
echo ""
echo "Step 4: Building Future.xcframework..."
cd "$BUILD_DIR/libXray"
gomobile bind -v \
    -target=ios,macos \
    -o "$BUILD_DIR/Future.xcframework" \
    -ldflags="-s -w" \
    .

echo "  Framework built."

# Step 5: Replace existing framework
echo ""
echo "Step 5: Replacing existing framework..."
rm -rf "$OUTPUT"
cp -R "$BUILD_DIR/Future.xcframework" "$OUTPUT"

echo ""
echo "=== Build complete ==="
echo "Framework: $OUTPUT"
echo ""
echo "Features included:"
echo "  - libXray (URI parsing, speed test, geo tools, iOS memory management)"
echo "  - Anti-DPI: ScatterConn, HeartbeatConn, record randomization"
echo "  - Anti-DPI: VLESS/VMess padding, SessionID jitter, post-handshake fragmentation"
echo "  - Apple packet flow integration (RegisterAppleNetworkInterface)"
echo ""
echo "To verify, open xNetFuture.xcworkspace in Xcode and build."

# Cleanup
rm -rf "$BUILD_DIR"
echo "Build directory cleaned up."
