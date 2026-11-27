#!/usr/bin/env bash
set -euo pipefail

export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${PATH:-}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d /tmp/bladerf-rdma.XXXXXX)"
SRV_LOG="$TMP_DIR/server.log"

SRV_PID=""
PRIOR_NETNS_MODE=""

cleanup() {
    local rc=$?
    set +e
    if [ -n "$SRV_PID" ]; then
        kill "$SRV_PID" 2>/dev/null
        wait "$SRV_PID" 2>/dev/null
    fi
    rdma link delete rxe_ci 2>/dev/null
    ip link del veth-rxe0 2>/dev/null
    if [ -n "$PRIOR_NETNS_MODE" ]; then
        rdma system set netns "$PRIOR_NETNS_MODE" 2>/dev/null
    fi
    if [ "$rc" -ne 0 ]; then
        echo "=== bladerf-remote log ===" >&2
        [ -f "$SRV_LOG" ] && cat "$SRV_LOG" >&2 || true
        dmesg | tail -n 30 >&2 || true
    fi
    rm -rf "$TMP_DIR"
    exit "$rc"
}
trap cleanup EXIT

if [ "$(id -u)" -ne 0 ]; then
    echo "run_rdma_netns.sh must run as root" >&2
    exit 1
fi

ulimit -l unlimited

modprobe crc32_generic 2>/dev/null || modprobe crc32 2>/dev/null || true
modprobe crc32-pclmul 2>/dev/null || true
modprobe veth 2>/dev/null || true
modprobe ib_core 2>/dev/null || true
modprobe rdma_cm 2>/dev/null || true
modprobe ib_uverbs 2>/dev/null || true
modprobe rdma_ucm 2>/dev/null || true
modprobe rdma_rxe

TARGET_IP="10.233.0.1"
PRIOR_NETNS_MODE="$(rdma system show 2>/dev/null | awk '/netns/ {print $2}' || true)"

rdma system set netns shared 2>/dev/null || true
rdma link delete rxe_ci 2>/dev/null || true
ip link del veth-rxe0 2>/dev/null || true

ip link set lo up
ip link add veth-rxe0 type veth peer name veth-rxe1
ip addr add "${TARGET_IP}/24" dev veth-rxe0
ip link set veth-rxe0 up
ip link set veth-rxe1 up
ethtool -K veth-rxe0 rx off tx off 2>/dev/null || ethtool -K veth-rxe0 tx off 2>/dev/null || true
ethtool -K veth-rxe1 rx off tx off 2>/dev/null || ethtool -K veth-rxe1 tx off 2>/dev/null || true
rdma link add rxe_ci type rxe netdev veth-rxe0

echo "=== SoftRoCE RDMA link ==="
rdma link show

"$ROOT/build/dist/bin/bladerf-remote" -l "${TARGET_IP}:5555" --sim >"$SRV_LOG" 2>&1 &
SRV_PID=$!

READY=0
for _ in $(seq 1 60); do
    if grep -q "listening RDMA (librdmacm) on ${TARGET_IP}:5556" "$SRV_LOG" 2>/dev/null && \
       grep -q "listening TCP on ${TARGET_IP}:5555" "$SRV_LOG" 2>/dev/null; then
        READY=1
        break
    fi
    if ! kill -0 "$SRV_PID" 2>/dev/null; then
        echo "bladerf-remote exited prematurely" >&2
        exit 1
    fi
    sleep 0.1
done

if [ "$READY" -ne 1 ]; then
    echo "bladerf-remote did not start RDMA + TCP listeners cleanly" >&2
    exit 1
fi

echo "=== Running Go live RDMA integration tests ==="
BLADERF_TEST_RDMA_ADDR="${TARGET_IP}:5555" "$ROOT/build/dist/bin/bladerf-rdma.test" -test.v -test.run '^TestLiveRDMA' -test.timeout 30s

echo "Live RDMA tests completed successfully"
