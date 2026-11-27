#!/usr/bin/env bash
set -euo pipefail

export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${PATH:-}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [ ! -f "$ROOT/build/dist/lib/libbladeRF.a" ] || \
   [ ! -x "$ROOT/build/dist/bin/bladerf-remote" ]; then
    make -C "$ROOT"
fi

echo "Building Go live RDMA test binary..."
(cd "$ROOT" && go test -tags=rdma_live -c -o "$ROOT/build/dist/bin/bladerf-rdma.test" .)

if [ "${FORCE_QEMU_RDMA:-0}" != "1" ] && \
   command -v ip >/dev/null 2>&1 && \
   command -v rdma >/dev/null 2>&1 && \
   command -v ethtool >/dev/null 2>&1; then
    if [ "$(id -u)" -eq 0 ] && modprobe -n rdma_rxe >/dev/null 2>&1; then
        exec "$ROOT/scripts/run_rdma_netns.sh"
    elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1 && sudo modprobe -n rdma_rxe >/dev/null 2>&1; then
        exec sudo "$ROOT/scripts/run_rdma_netns.sh"
    fi
fi

echo "Host lacks passwordless kernel rdma_rxe access; booting ephemeral QEMU SoftRoCE VM..."

HOST_ARCH="$(uname -m)"
case "$HOST_ARCH" in
    x86_64)
        QEMU_BIN="qemu-system-x86_64"
        QEMU_MACH="-machine pc"
        CONSOLE="ttyS0"
        ;;
    aarch64|arm64)
        QEMU_BIN="qemu-system-aarch64"
        QEMU_MACH="-machine virt"
        CONSOLE="ttyAMA0"
        ;;
    *)
        echo "Unsupported host architecture for QEMU RDMA fallback: $HOST_ARCH" >&2
        exit 1
        ;;
esac

for tool in "$QEMU_BIN" cpio depmod; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "Required tool not found for QEMU RDMA test: $tool" >&2
        exit 1
    fi
done

if [ -r /dev/kvm ] && [ -w /dev/kvm ]; then
    QEMU_ACCEL="-accel kvm:tcg -cpu max"
else
    QEMU_ACCEL="-accel tcg -cpu max"
fi

KVER="$(uname -r)"
if ! apt-cache show "linux-modules-extra-${KVER}" >/dev/null 2>&1; then
    KVER="$(apt-cache depends linux-image-generic 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+-[0-9]+-generic' | head -n 1 || true)"
fi
if [ -z "$KVER" ]; then
    echo "Unable to resolve Ubuntu kernel package version for SoftRoCE VM" >&2
    exit 1
fi

VM_DIR="$ROOT/build/rdma-vm"
DEBS_DIR="$VM_DIR/debs"
MODROOT="$VM_DIR/modroot"
INITRD_DIR="$VM_DIR/initramfs"
INITRD_IMG="$VM_DIR/initramfs.cpio"
VM_LOG="$VM_DIR/vm.log"

mkdir -p "$DEBS_DIR" "$MODROOT"

if [ ! -f "$MODROOT/boot/vmlinuz-$KVER" ] || \
   [ -z "$(find "$MODROOT/lib/modules/$KVER" -name 'rdma_rxe.ko*' 2>/dev/null || true)" ]; then
    echo "Preparing kernel + SoftRoCE modules ($KVER) in $MODROOT..."
    rm -rf "$MODROOT"
    mkdir -p "$MODROOT/lib/modules/$KVER"

    NEED_PKGS=("linux-image-${KVER}" "linux-modules-extra-${KVER}")
    HOST_MODS="/lib/modules/${KVER}/kernel"
    if [ "$KVER" = "$(uname -r)" ] && [ -d "$HOST_MODS" ]; then
        cp -a "/lib/modules/${KVER}/." "$MODROOT/lib/modules/${KVER}/"
    else
        NEED_PKGS+=("linux-modules-${KVER}")
    fi

    (cd "$DEBS_DIR" && apt-get download "${NEED_PKGS[@]}")
    for pkg in "${NEED_PKGS[@]}"; do
        for deb in "$DEBS_DIR/${pkg}_"*.deb; do
            if [ -f "$deb" ]; then
                dpkg-deb -x "$deb" "$MODROOT"
            fi
        done
    done

    if [ -d "$MODROOT/usr/lib/modules/${KVER}" ]; then
        cp -a "$MODROOT/usr/lib/modules/${KVER}/." "$MODROOT/lib/modules/${KVER}/"
    fi
    rm -f "$MODROOT/lib/modules/${KVER}"/modules.*
    depmod -b "$MODROOT" "$KVER"
fi

VMLINUZ="$MODROOT/boot/vmlinuz-$KVER"
if [ ! -f "$VMLINUZ" ]; then
    VMLINUZ="$(find "$MODROOT/boot" -maxdepth 1 -name 'vmlinuz-*' | head -n 1 || true)"
fi
if [ -z "$VMLINUZ" ] || [ ! -f "$VMLINUZ" ]; then
    echo "Kernel image not found in $MODROOT/boot" >&2
    exit 1
fi
if [ "$HOST_ARCH" = "aarch64" ] || [ "$HOST_ARCH" = "arm64" ]; then
    if gzip -t "$VMLINUZ" >/dev/null 2>&1; then
        gzip -dc "$VMLINUZ" > "$VM_DIR/Image"
        VMLINUZ="$VM_DIR/Image"
    fi
fi

BUSYBOX_BIN=""
if command -v busybox >/dev/null 2>&1 && ldd "$(command -v busybox)" 2>&1 | grep -qE 'not a dynamic executable|statically linked'; then
    BUSYBOX_BIN="$(command -v busybox)"
else
    if [ ! -x "$VM_DIR/bbroot/bin/busybox" ]; then
        mkdir -p "$VM_DIR/bbroot"
        (cd "$DEBS_DIR" && apt-get download busybox-static)
        for deb in "$DEBS_DIR"/busybox-static_*.deb; do
            dpkg-deb -x "$deb" "$VM_DIR/bbroot"
        done
    fi
    if [ -x "$VM_DIR/bbroot/bin/busybox" ]; then
        BUSYBOX_BIN="$VM_DIR/bbroot/bin/busybox"
    elif [ -x "$VM_DIR/bbroot/usr/bin/busybox" ]; then
        BUSYBOX_BIN="$VM_DIR/bbroot/usr/bin/busybox"
    fi
fi
if [ -z "$BUSYBOX_BIN" ]; then
    echo "Static busybox binary not found" >&2
    exit 1
fi

rm -rf "$INITRD_DIR"
mkdir -p "$INITRD_DIR/bin" "$INITRD_DIR/modules" "$INITRD_DIR/proc" "$INITRD_DIR/sys" "$INITRD_DIR/dev" "$INITRD_DIR/mnt/host" "$INITRD_DIR/mnt/root"
cp "$BUSYBOX_BIN" "$INITRD_DIR/bin/busybox"
chmod +x "$INITRD_DIR/bin/busybox"
for cmd in sh mount insmod mkdir ln chmod chroot poweroff cp uname readlink sleep cat umount; do
    ln -sf busybox "$INITRD_DIR/bin/$cmd"
done

stage_ko() {
    local mod="$1"
    local src
    src="$(find "$MODROOT/lib/modules/$KVER" -name "${mod}.ko*" 2>/dev/null | head -n 1 || true)"
    if [ -z "$src" ]; then
        return 0
    fi
    local dst="$INITRD_DIR/modules/${mod}.ko"
    case "$src" in
        *.zst) zstd -dqf -o "$dst" "$src" ;;
        *.gz)  gzip -dc "$src" > "$dst" ;;
        *.xz)  xz -dc "$src" > "$dst" ;;
        *)     cp "$src" "$dst" ;;
    esac
}

for m in virtio virtio_ring virtio_pci_legacy_dev virtio_pci_modern_dev virtio_pci virtio_mmio netfs 9pnet 9pnet_virtio 9p; do
    stage_ko "$m"
done

cat > "$INITRD_DIR/init" <<INIT_EOF
#!/bin/busybox sh
set -eu

export PATH="/bin"
KVER="${KVER}"
ROOT_WS="${ROOT}"
MODROOT_WS="${MODROOT}"

mkdir -p /proc /sys /dev /mnt/host /mnt/root
mount -t proc proc /proc
mount -t sysfs sysfs /sys
mount -t devtmpfs devtmpfs /dev

for mod in virtio virtio_ring virtio_pci_legacy_dev virtio_pci_modern_dev virtio_pci virtio_mmio netfs 9pnet 9pnet_virtio 9p; do
    if [ -f "/modules/\${mod}.ko" ]; then
        insmod "/modules/\${mod}.ko" 2>/dev/null || true
    fi
done

mount -t 9p -o trans=virtio,version=9p2000.L,ro,msize=1048576 hostroot /mnt/host
mount -t tmpfs tmpfs /mnt/root

mkdir -p /mnt/root/proc /mnt/root/sys /mnt/root/dev /mnt/root/tmp /mnt/root/run /mnt/root/var /mnt/root/etc /mnt/root/root /mnt/root/mnt
chmod 1777 /mnt/root/tmp

mount --move /proc /mnt/root/proc
mount --move /sys /mnt/root/sys
mount --move /dev /mnt/root/dev
mkdir -p /mnt/root/dev/pts /mnt/root/dev/shm
mount -t devpts devpts /mnt/root/dev/pts
mount -t tmpfs tmpfs /mnt/root/dev/shm

mount -t tmpfs tmpfs /mnt/root/run
mount --make-rshared /mnt/root/run
mkdir -p /mnt/root/run/netns /mnt/root/run/lock
ln -s /run /mnt/root/var/run
mkdir -p /mnt/root/var/tmp
chmod 1777 /mnt/root/var/tmp

for d in bin sbin usr lib lib64; do
    if [ -L "/mnt/host/\$d" ]; then
        ln -s "\$(readlink "/mnt/host/\$d")" "/mnt/root/\$d"
    elif [ -d "/mnt/host/\$d" ]; then
        mkdir -p "/mnt/root/\$d"
        mount --bind "/mnt/host/\$d" "/mnt/root/\$d"
    fi
done

if [ -d /mnt/root/usr/lib/modules ]; then
    mount -t tmpfs tmpfs /mnt/root/usr/lib/modules
    mkdir -p "/mnt/root/usr/lib/modules/\${KVER}"
    mount --bind "/mnt/host\${MODROOT_WS}/lib/modules/\${KVER}" "/mnt/root/usr/lib/modules/\${KVER}"
elif [ -d /mnt/root/lib/modules ]; then
    mount -t tmpfs tmpfs /mnt/root/lib/modules
    mkdir -p "/mnt/root/lib/modules/\${KVER}"
    mount --bind "/mnt/host\${MODROOT_WS}/lib/modules/\${KVER}" "/mnt/root/lib/modules/\${KVER}"
fi

for item in ld.so.cache ld.so.conf ld.so.conf.d libibverbs.d rdma iproute2 alternatives modprobe.d depmod.d passwd group hosts nsswitch.conf protocols services; do
    if [ -e "/mnt/host/etc/\$item" ]; then
        cp -a "/mnt/host/etc/\$item" /mnt/root/etc/ 2>/dev/null || true
    fi
done

mkdir -p "/mnt/root\${ROOT_WS}"
mount --bind "/mnt/host\${ROOT_WS}" "/mnt/root\${ROOT_WS}"

set +e
chroot /mnt/root /usr/bin/env PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin" /bin/bash "\${ROOT_WS}/scripts/run_rdma_netns.sh"
rc=\$?
set -e

echo "RDMA_VM_EXIT_CODE=\$rc"
sleep 0.1
poweroff -f
INIT_EOF
chmod +x "$INITRD_DIR/init"

(cd "$INITRD_DIR" && find . -print0 | cpio --null -o --format=newc --quiet > "$INITRD_IMG")

echo "Booting QEMU SoftRoCE VM ($VMLINUZ)..."
set +e
timeout 180 "$QEMU_BIN" \
    $QEMU_MACH \
    $QEMU_ACCEL \
    -m 1024M \
    -smp 2 \
    -display none \
    -monitor none \
    -serial stdio \
    -no-reboot \
    -kernel "$VMLINUZ" \
    -initrd "$INITRD_IMG" \
    -append "console=${CONSOLE} panic=1 loglevel=4 rdinit=/init" \
    -virtfs local,path=/,mount_tag=hostroot,security_model=none,readonly=on \
    </dev/null 2>&1 | tee "$VM_LOG"
qemu_rc=${PIPESTATUS[0]}
set -e

if [ "$qemu_rc" -ne 0 ] || ! tr -d '\r' < "$VM_LOG" | grep -q '^RDMA_VM_EXIT_CODE=0$'; then
    echo "QEMU SoftRoCE VM RDMA test failed (qemu_rc=$qemu_rc)" >&2
    exit 1
fi
