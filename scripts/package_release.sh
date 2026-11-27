#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${1:-dev}"
RAW_ARCH="${2:-$(dpkg --print-architecture 2>/dev/null || uname -m)}"

case "$RAW_ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) ARCH="$RAW_ARCH" ;;
esac

if [ ! -f "$ROOT/build/dist/lib/libbladeRF.a" ] || [ ! -f "$ROOT/build/dist/lib/libwhisper.a" ] || [ ! -x "$ROOT/build/dist/bin/bladerf-remote" ]; then
    "$ROOT/scripts/build_libbladerf.sh"
fi

if [ -d "$ROOT/build/deps/pkgconfig" ]; then
    export PKG_CONFIG_PATH="$ROOT/build/deps/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
fi

PKG_NAME="go-bladerf-${VERSION}-linux-${ARCH}"
RELEASE_DIR="$ROOT/build/release"
STAGE_DIR="$RELEASE_DIR/$PKG_NAME"

rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR/bin" "$STAGE_DIR/lib" "$STAGE_DIR/include"

cp -a "$ROOT/build/dist/bin/bladerf-remote" "$STAGE_DIR/bin/bladerf-remote"
if command -v strip >/dev/null 2>&1; then
    strip --strip-unneeded "$STAGE_DIR/bin/bladerf-remote"
fi
ln -sf bladerf-remote "$STAGE_DIR/bin/bladeRF-remote"
ln -sf bladerf-remote "$STAGE_DIR/bin/bladerf-tcp"

(cd "$ROOT" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$STAGE_DIR/bin/bladerf-mock" ./cmd/bladerf-mock)

cp -a "$ROOT/build/dist/lib/libbladeRF.a" "$STAGE_DIR/lib/"
cp -a "$ROOT/build/dist/lib/libportaudio.a" "$STAGE_DIR/lib/"
cp -a "$ROOT/build/dist/lib/libwhisper.a" "$STAGE_DIR/lib/"
cp -a "$ROOT/build/dist/include"/*.h "$STAGE_DIR/include/"
cp -a "$ROOT/README.md" "$STAGE_DIR/README.md"

TARBALL="${RELEASE_DIR}/${PKG_NAME}.tar.gz"
STANDALONE_MOCK="${RELEASE_DIR}/bladerf-mock-${VERSION}-linux-${ARCH}"

rm -f "$TARBALL" "$STANDALONE_MOCK"
tar -C "$RELEASE_DIR" -czf "$TARBALL" "$PKG_NAME"
cp -a "$STAGE_DIR/bin/bladerf-mock" "$STANDALONE_MOCK"

(cd "$RELEASE_DIR" && sha256sum "${PKG_NAME}.tar.gz" "bladerf-mock-${VERSION}-linux-${ARCH}" > "SHA256SUMS-${ARCH}.txt")

echo "Packaged release artifacts in $RELEASE_DIR:"
ls -lh "$TARBALL" "$STANDALONE_MOCK" "$RELEASE_DIR/SHA256SUMS-${ARCH}.txt"
