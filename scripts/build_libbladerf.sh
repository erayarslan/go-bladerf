#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
THIRD_PARTY="$ROOT/thirdparty/bladeRF"
PATCH_FILE="$ROOT/patches/0001-remote-rdma-tcp-backend.patch"
BUILD_SRC="$ROOT/build/bladeRF-src"
BUILD_CMAKE="$ROOT/build/cmake"
DIST_DIR="$ROOT/build/dist"

DEPS_DIR="$ROOT/build/deps"
SRC_CACHE="$DEPS_DIR/src"
STATIC_PREFIX="$DEPS_DIR/static"
SYSROOT="$DEPS_DIR/sysroot"
DEPLIB="$DEPS_DIR/lib"
PKGCFG_DIR="$DEPS_DIR/pkgconfig"
NPROC="$(nproc 2>/dev/null || echo 4)"
MULTIARCH="$(gcc -print-multiarch 2>/dev/null || echo "x86_64-linux-gnu")"

LIBUSB_VER="1.0.27"
LIBUSB_URL="https://github.com/libusb/libusb/releases/download/v${LIBUSB_VER}/libusb-${LIBUSB_VER}.tar.bz2"
PORTAUDIO_VER="19.7.0"
PORTAUDIO_URL="https://github.com/PortAudio/portaudio/archive/refs/tags/v${PORTAUDIO_VER}.tar.gz"
WHISPER_VER="1.7.6"
WHISPER_URL="https://github.com/ggml-org/whisper.cpp/archive/refs/tags/v${WHISPER_VER}.tar.gz"

if [ ! -f "$THIRD_PARTY/host/CMakeLists.txt" ]; then
    echo "Initializing thirdparty/bladeRF submodule..."
    git -C "$ROOT" submodule update --init thirdparty/bladeRF
    git -C "$THIRD_PARTY" submodule update --init thirdparty/analogdevicesinc/no-OS
fi

mkdir -p "$DEPS_DIR" "$DEPS_DIR/bin" "$SRC_CACHE" "$STATIC_PREFIX" "$SYSROOT" "$DEPLIB" "$PKGCFG_DIR"

# 1. Build static PIC libusb-1.0.a using Linux kernel netlink + usbfs (zero libudev dependency).
if [ ! -f "$STATIC_PREFIX/lib/libusb-1.0.a" ] || [ ! -f "$STATIC_PREFIX/include/libusb-1.0/libusb.h" ]; then
    echo "Building static PIC libusb-${LIBUSB_VER} (--disable-udev)..."
    LIBUSB_TBZ="$SRC_CACHE/libusb-${LIBUSB_VER}.tar.bz2"
    if [ ! -s "$LIBUSB_TBZ" ]; then
        curl -fsSL "$LIBUSB_URL" -o "$LIBUSB_TBZ"
    fi
    rm -rf "$DEPS_DIR/build-libusb"
    mkdir -p "$DEPS_DIR/build-libusb"
    tar -xjf "$LIBUSB_TBZ" -C "$DEPS_DIR/build-libusb"
    (
        cd "$DEPS_DIR/build-libusb/libusb-${LIBUSB_VER}"
        CFLAGS="-O2 -fPIC" ./configure \
            --prefix="$STATIC_PREFIX" \
            --enable-static \
            --disable-shared \
            --with-pic \
            --disable-udev
        make -j"$NPROC"
        make install
    )
fi

# 2. Ensure ALSA headers and link-time libasound.so symlink exist for PortAudio PA_ALSA_DYNAMIC detection.
ALSA_INC=""
ALSA_LDFLAGS="-L$DEPLIB"
if [ ! -f /usr/include/alsa/asoundlib.h ] && [ ! -f "$SYSROOT/usr/include/alsa/asoundlib.h" ]; then
    echo "Downloading ALSA headers into build/deps/sysroot..."
    mkdir -p "$DEPS_DIR/pa-debs"
    (
        cd "$DEPS_DIR/pa-debs"
        apt-get download libasound2-dev
        if ! ls "/usr/lib/$MULTIARCH"/libasound.so.2* "/lib/$MULTIARCH"/libasound.so.2* >/dev/null 2>&1; then
            apt-get download libasound2t64 2>/dev/null || apt-get download libasound2
        fi
    )
    for deb in "$DEPS_DIR/pa-debs"/*.deb; do
        dpkg-deb -x "$deb" "$SYSROOT"
    done
fi

if [ -f "$SYSROOT/usr/include/alsa/asoundlib.h" ]; then
    ALSA_INC="-I$SYSROOT/usr/include"
fi

ASND_SO=""
for cand in \
    "/usr/lib/$MULTIARCH/libasound.so.2" \
    "/lib/$MULTIARCH/libasound.so.2" \
    "$SYSROOT/usr/lib/$MULTIARCH/libasound.so.2"; do
    if [ -e "$cand" ]; then
        ASND_SO="$cand"
        break
    fi
done
if [ -n "$ASND_SO" ]; then
    ln -sf "$ASND_SO" "$DEPLIB/libasound.so"
    ALSA_DIR="$(dirname "$ASND_SO")"
    ALSA_LDFLAGS="$ALSA_LDFLAGS -L$ALSA_DIR -Wl,-rpath-link,$ALSA_DIR"
fi

# 3. Build static PIC libportaudio.a (PA_ALSA_DYNAMIC dlopen("libasound.so.2") + OSS, no JACK).
if [ ! -f "$STATIC_PREFIX/lib/libportaudio.a" ] || [ ! -f "$STATIC_PREFIX/include/portaudio.h" ]; then
    echo "Building static PIC PortAudio-${PORTAUDIO_VER} (PA_ALSA_DYNAMIC=1, no JACK)..."
    PA_TGZ="$SRC_CACHE/portaudio-${PORTAUDIO_VER}.tar.gz"
    if [ ! -s "$PA_TGZ" ]; then
        curl -fsSL "$PORTAUDIO_URL" -o "$PA_TGZ"
    fi
    rm -rf "$DEPS_DIR/build-portaudio"
    mkdir -p "$DEPS_DIR/build-portaudio"
    tar -xzf "$PA_TGZ" -C "$DEPS_DIR/build-portaudio"
    (
        cd "$DEPS_DIR/build-portaudio/portaudio-${PORTAUDIO_VER}"
        CPPFLAGS="$ALSA_INC" \
        CFLAGS="-O2 -fPIC $ALSA_INC -DPA_ALSA_DYNAMIC=1 -DPA_ALSA_PATHNAME=\\\"libasound.so.2\\\"" \
        LDFLAGS="$ALSA_LDFLAGS" \
        LIBS="-ldl" \
        ./configure \
            --prefix="$STATIC_PREFIX" \
            --enable-static \
            --disable-shared \
            --with-pic \
            --with-alsa \
            --with-oss \
            --without-jack
        grep -q 'PA_USE_ALSA' Makefile
        make -j1 lib/libportaudio.la
        make -j1 install
    )
    if nm -u "$STATIC_PREFIX/lib/libportaudio.a" | grep -qE ' (snd_|jack_)'; then
        echo "ERROR: libportaudio.a unexpectedly has direct ALSA/JACK symbol imports" >&2
        exit 1
    fi
fi

# 4. Patch and build libbladeRF_static + bladeRF-remote with RPATH disabled.
rm -rf "$BUILD_SRC" "$BUILD_CMAKE"
mkdir -p "$BUILD_SRC" "$DIST_DIR/include" "$DIST_DIR/lib" "$DIST_DIR/bin"

rsync -a --delete --exclude='.git' "$THIRD_PARTY/" "$BUILD_SRC/"
patch -p1 -d "$BUILD_SRC" < "$PATCH_FILE"

cmake -S "$BUILD_SRC/host" -B "$BUILD_CMAKE" \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_INSTALL_PREFIX="$DIST_DIR" \
    -DCMAKE_INSTALL_LIBDIR=lib \
    -DCMAKE_SKIP_RPATH=ON \
    -DTREAT_WARNINGS_AS_ERRORS=OFF \
    -DBUILD_DOCUMENTATION=OFF \
    -DBUILD_BLADERF_CLI=OFF \
    -DBUILD_BLADERF_UPDATE=OFF \
    -DBUILD_BLADERF_POWER=OFF \
    -DBUILD_BLADERF_FSK=OFF \
    -DTEST_LIBBLADERF=OFF \
    -DENABLE_BACKEND_REMOTE=ON \
    -DENABLE_BACKEND_RDMA=ON \
    -DLIBUSB_FOUND=TRUE \
    -DLIBUSB_HEADER_FILE="$STATIC_PREFIX/include/libusb-1.0/libusb.h" \
    -DLIBUSB_INCLUDE_DIRS="$STATIC_PREFIX/include/libusb-1.0" \
    -DLIBUSB_LIBRARIES="$STATIC_PREFIX/lib/libusb-1.0.a" \
    -Dusb_LIBRARY="$STATIC_PREFIX/lib/libusb-1.0.a" \
    -DLIBUSB_SKIP_VERSION_CHECK=ON

cmake --build "$BUILD_CMAKE" --target libbladerf_static bladeRF-remote -j"$NPROC"

cp -a "$BUILD_SRC/host/libraries/libbladeRF/include/libbladeRF.h" "$DIST_DIR/include/"
cp -a "$BUILD_SRC/host/libraries/libbladeRF/include/bladeRF1.h" "$DIST_DIR/include/"
cp -a "$BUILD_SRC/host/libraries/libbladeRF/include/bladeRF2.h" "$DIST_DIR/include/"
cp -a "$STATIC_PREFIX/include/portaudio.h" "$DIST_DIR/include/"
if [ -f "$STATIC_PREFIX/include/pa_linux_alsa.h" ]; then
    cp -a "$STATIC_PREFIX/include/pa_linux_alsa.h" "$DIST_DIR/include/"
fi

# Merge libbladeRF + libad936x + libusb-1.0 into a single self-contained libbladeRF.a.
rm -f "$DIST_DIR/lib"/libbladeRF.so* "$DIST_DIR/lib/libbladeRF.a"
ar -M <<EOF
CREATE $DIST_DIR/lib/libbladeRF.a
ADDLIB $BUILD_CMAKE/output/libbladeRF.a
ADDLIB $BUILD_CMAKE/output/libad936x.a
ADDLIB $STATIC_PREFIX/lib/libusb-1.0.a
SAVE
END
EOF
ranlib "$DIST_DIR/lib/libbladeRF.a"
cp -a "$STATIC_PREFIX/lib/libportaudio.a" "$DIST_DIR/lib/libportaudio.a"
cp -a "$STATIC_PREFIX/lib/libusb-1.0.a" "$DIST_DIR/lib/libusb-1.0.a"

# Always point portaudio-2.0.pc to the static PIC archive with dynamic ALSA dlopen (no JACK/asound NEEDED).
cat > "$PKGCFG_DIR/portaudio-2.0.pc" <<EOF
prefix=${DIST_DIR}
exec_prefix=${DIST_DIR}
libdir=${DIST_DIR}/lib
includedir=${DIST_DIR}/include

Name: PortAudio
Description: Portable audio I/O (static with runtime ALSA dlopen)
Version: ${PORTAUDIO_VER}
Libs: ${DIST_DIR}/lib/libportaudio.a -ldl -lm -lpthread
Cflags: -I${DIST_DIR}/include -pthread
EOF

# 5. Build static PIC whisper.cpp (ggml-cpu + ggml-vulkan) with a runtime Vulkan dlopen shim (zero libvulkan/libstdc++ NEEDED).
if [ ! -f "$STATIC_PREFIX/lib/libwhisper_combined.a" ] || [ ! -f "$STATIC_PREFIX/include/whisper.h" ]; then
    echo "Preparing Vulkan build headers & glslc shader compiler for whisper.cpp-${WHISPER_VER}..."
    mkdir -p "$DEPS_DIR/vk-debs"

    VK_INC_DIR=""
    if [ -f /usr/include/vulkan/vulkan.hpp ]; then
        VK_INC_DIR="/usr/include"
    elif [ -f "$SYSROOT/usr/include/vulkan/vulkan.hpp" ]; then
        VK_INC_DIR="$SYSROOT/usr/include"
    else
        (
            cd "$DEPS_DIR/vk-debs"
            apt-get download libvulkan-dev
            if ! ls "/usr/lib/$MULTIARCH"/libvulkan.so.1* "/lib/$MULTIARCH"/libvulkan.so.1* >/dev/null 2>&1; then
                apt-get download libvulkan1
            fi
        )
        for deb in "$DEPS_DIR/vk-debs"/*.deb; do
            dpkg-deb -x "$deb" "$SYSROOT"
        done
        VK_INC_DIR="$SYSROOT/usr/include"
    fi

    GLSLC_BIN=""
    if command -v glslc >/dev/null 2>&1; then
        GLSLC_BIN="$(command -v glslc)"
    elif [ -x "$SYSROOT/usr/bin/glslc" ]; then
        GLSLC_BIN="$DEPS_DIR/bin/glslc"
    else
        (
            cd "$DEPS_DIR/vk-debs"
            apt-get download glslc
            SHADERC_PKG="$(apt-cache depends glslc | awk '/Depends: libshaderc/ {print $2; exit}')"
            if [ -n "$SHADERC_PKG" ]; then
                apt-get download "$SHADERC_PKG"
            fi
        )
        for deb in "$DEPS_DIR/vk-debs"/*.deb; do
            dpkg-deb -x "$deb" "$SYSROOT"
        done
        GLSLC_BIN="$DEPS_DIR/bin/glslc"
    fi

    if [ "$GLSLC_BIN" = "$DEPS_DIR/bin/glslc" ]; then
        cat > "$GLSLC_BIN" <<EOF
#!/usr/bin/env bash
export LD_LIBRARY_PATH="${SYSROOT}/usr/lib/${MULTIARCH}:${SYSROOT}/lib/${MULTIARCH}:${SYSROOT}/usr/lib\${LD_LIBRARY_PATH:+:\$LD_LIBRARY_PATH}"
exec "${SYSROOT}/usr/bin/glslc" "\$@"
EOF
        chmod +x "$GLSLC_BIN"
    fi

    VK_SO=""
    for cand in \
        "/usr/lib/$MULTIARCH/libvulkan.so.1" \
        "/lib/$MULTIARCH/libvulkan.so.1" \
        "$SYSROOT/usr/lib/$MULTIARCH/libvulkan.so.1"; do
        if [ -e "$cand" ]; then
            VK_SO="$cand"
            break
        fi
    done
    if [ -z "$VK_SO" ]; then
        echo "ERROR: libvulkan.so.1 not found in system or sysroot for CMake Vulkan detection" >&2
        exit 1
    fi
    ln -sf "$VK_SO" "$DEPLIB/libvulkan.so"

    echo "Building static PIC whisper.cpp-${WHISPER_VER} (Vulkan GPU + portable CPU fallback)..."
    WHISPER_TGZ="$SRC_CACHE/whisper-${WHISPER_VER}.tar.gz"
    if [ ! -s "$WHISPER_TGZ" ]; then
        curl -fsSL "$WHISPER_URL" -o "$WHISPER_TGZ"
    fi
    rm -rf "$DEPS_DIR/build-whisper"
    mkdir -p "$DEPS_DIR/build-whisper"
    tar -xzf "$WHISPER_TGZ" -C "$DEPS_DIR/build-whisper"

    WHISPER_SRC="$DEPS_DIR/build-whisper/whisper.cpp-${WHISPER_VER}"
    WHISPER_BUILD="$DEPS_DIR/build-whisper/build"

    SIMD_FLAGS=(-DGGML_NATIVE=OFF)
    case "$(uname -m)" in
        x86_64|amd64)
            SIMD_FLAGS+=(
                -DGGML_AVX512=OFF
                -DGGML_AVX512_VBMI=OFF
                -DGGML_AVX512_VNNI=OFF
                -DGGML_AVX512_BF16=OFF
                -DGGML_SSE42=ON
                -DGGML_AVX=ON
                -DGGML_AVX2=ON
                -DGGML_FMA=ON
                -DGGML_F16C=ON
            )
            ;;
    esac

    cmake -S "$WHISPER_SRC" -B "$WHISPER_BUILD" \
        -DCMAKE_BUILD_TYPE=Release \
        -DBUILD_SHARED_LIBS=OFF \
        -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
        -DCMAKE_SKIP_RPATH=ON \
        -DWHISPER_BUILD_TESTS=OFF \
        -DWHISPER_BUILD_EXAMPLES=OFF \
        -DWHISPER_BUILD_SERVER=OFF \
        -DGGML_OPENMP=OFF \
        -DGGML_VULKAN=ON \
        -DVulkan_INCLUDE_DIR="$VK_INC_DIR" \
        -DVulkan_LIBRARY="$DEPLIB/libvulkan.so" \
        -DVulkan_GLSLC_EXECUTABLE="$GLSLC_BIN" \
        "${SIMD_FLAGS[@]}"

    cmake --build "$WHISPER_BUILD" -j"$NPROC"

    LIB_WHISPER_A="$(find "$WHISPER_BUILD" -name 'libwhisper.a' | head -n1)"
    LIB_GGML_A="$(find "$WHISPER_BUILD" -name 'libggml.a' | head -n1)"
    LIB_GGML_BASE_A="$(find "$WHISPER_BUILD" -name 'libggml-base.a' | head -n1)"
    LIB_GGML_CPU_A="$(find "$WHISPER_BUILD" -name 'libggml-cpu.a' | head -n1)"
    LIB_GGML_VK_A="$(find "$WHISPER_BUILD" -name 'libggml-vulkan.a' | head -n1)"
    for libf in "$LIB_WHISPER_A" "$LIB_GGML_A" "$LIB_GGML_BASE_A" "$LIB_GGML_CPU_A" "$LIB_GGML_VK_A"; do
        if [ -z "$libf" ] || [ ! -f "$libf" ]; then
            echo "ERROR: Missing built whisper/ggml static archive ($libf)" >&2
            exit 1
        fi
    done

    VK_SHIM_C="$DEPS_DIR/build-whisper/vulkan_dlopen_shim.c"
    VK_SHIM_O="$DEPS_DIR/build-whisper/vulkan_dlopen_shim.o"
    python3 - "$LIB_GGML_VK_A" "$VK_INC_DIR/vulkan/vulkan_core.h" "$VK_SHIM_C" <<'PYEOF'
import sys, re, subprocess

ggml_vk_a, vk_hdr_path, out_c = sys.argv[1], sys.argv[2], sys.argv[3]
nm_out = subprocess.check_output(["nm", "-u", ggml_vk_a], text=True)
needed = sorted({
    line.strip().split()[-1]
    for line in nm_out.splitlines()
    if line.strip() and re.search(r'\bU\s+(vk[A-Z][A-Za-z0-9_]*)$', line.strip())
})

with open(vk_hdr_path, "r", encoding="utf-8") as f:
    hdr = f.read()

# Strip comments & collapse whitespace for clean VKAPI_ATTR ... VKAPI_CALL vk...(...) parsing.
hdr_clean = re.sub(r'/\*.*?\*/', ' ', hdr, flags=re.S)
hdr_clean = re.sub(r'//[^\n]*', ' ', hdr_clean)
sigs = {}
for m in re.finditer(r'VKAPI_ATTR\s+(.*?)\s+VKAPI_CALL\s+(vk[A-Za-z0-9_]+)\s*\(([^;{}]*?)\)\s*;', hdr_clean, flags=re.S):
    ret = " ".join(m.group(1).split())
    name = m.group(2).strip()
    raw_args = " ".join(m.group(3).split())
    sigs[name] = (ret, raw_args)

def parse_arg_names(raw_args):
    if raw_args == "void" or raw_args == "":
        return []
    names = []
    for part in raw_args.split(","):
        p = part.strip()
        p = re.sub(r'\[[^\]]*\]', '', p).strip()
        tok = p.split()[-1].lstrip("*")
        names.append(tok)
    return names

missing = [s for s in needed if s not in sigs]
if missing:
    raise SystemExit("ERROR: Missing Vulkan signatures in vulkan_core.h for: " + ", ".join(missing))

lines = []
lines.append("""#define VK_NO_PROTOTYPES
#include <vulkan/vulkan.h>
#include <dlfcn.h>
#include <pthread.h>
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <stdio.h>

static pthread_once_t g_vk_once = PTHREAD_ONCE_INIT;
static void *g_vk_lib = NULL;
static PFN_vkGetInstanceProcAddr g_vkGetInstanceProcAddr = NULL;
static VkInstance g_last_instance = VK_NULL_HANDLE;
static int g_vk_dummy_mode = 0;

static void init_vulkan_dl(void) {
    g_vk_lib = dlopen("libvulkan.so.1", RTLD_NOW | RTLD_LOCAL);
    if (!g_vk_lib) {
        g_vk_lib = dlopen("libvulkan.so", RTLD_NOW | RTLD_LOCAL);
    }
    if (g_vk_lib) {
        g_vkGetInstanceProcAddr = (PFN_vkGetInstanceProcAddr)dlsym(g_vk_lib, "vkGetInstanceProcAddr");
    }
    if (!g_vk_lib || !g_vkGetInstanceProcAddr) {
        g_vk_dummy_mode = 1;
    }
}

static void *resolve_vk_sym(const char *name) {
    pthread_once(&g_vk_once, init_vulkan_dl);
    if (!g_vk_lib) {
        return NULL;
    }
    void *sym = dlsym(g_vk_lib, name);
    if (!sym && g_vkGetInstanceProcAddr && g_last_instance != VK_NULL_HANDLE && g_last_instance != (VkInstance)(uintptr_t)0x1) {
        sym = (void *)g_vkGetInstanceProcAddr(g_last_instance, name);
    }
    return sym;
}

/* Safely checks whether at least one real non-CPU Vulkan GPU (discrete/integrated/virtual) is usable. */
int bladerf_vulkan_probe_gpu(char *name_out, size_t name_cap) {
    if (name_out && name_cap > 0) {
        name_out[0] = '\\0';
    }
    pthread_once(&g_vk_once, init_vulkan_dl);
    if (!g_vk_lib || !g_vkGetInstanceProcAddr) {
        return 0;
    }
    PFN_vkEnumerateInstanceVersion fn_enum_ver = (PFN_vkEnumerateInstanceVersion)dlsym(g_vk_lib, "vkEnumerateInstanceVersion");
    uint32_t api_ver = VK_MAKE_API_VERSION(0, 1, 0, 0);
    if (fn_enum_ver) {
        if (fn_enum_ver(&api_ver) != VK_SUCCESS || api_ver < VK_MAKE_API_VERSION(0, 1, 2, 0)) {
            return 0;
        }
    } else {
        return 0;
    }
    PFN_vkCreateInstance fn_create = (PFN_vkCreateInstance)g_vkGetInstanceProcAddr(VK_NULL_HANDLE, "vkCreateInstance");
    if (!fn_create) {
        fn_create = (PFN_vkCreateInstance)dlsym(g_vk_lib, "vkCreateInstance");
    }
    if (!fn_create) {
        return 0;
    }
    VkApplicationInfo app;
    memset(&app, 0, sizeof(app));
    app.sType = VK_STRUCTURE_TYPE_APPLICATION_INFO;
    app.pApplicationName = "bladerf-whisper-probe";
    app.applicationVersion = 1;
    app.pEngineName = "ggml-vulkan";
    app.engineVersion = 1;
    app.apiVersion = api_ver;

    VkInstanceCreateInfo ci;
    memset(&ci, 0, sizeof(ci));
    ci.sType = VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO;
    ci.pApplicationInfo = &app;

    VkInstance inst = VK_NULL_HANDLE;
    if (fn_create(&ci, NULL, &inst) != VK_SUCCESS || inst == VK_NULL_HANDLE) {
        return 0;
    }

    PFN_vkDestroyInstance fn_destroy = (PFN_vkDestroyInstance)g_vkGetInstanceProcAddr(inst, "vkDestroyInstance");
    PFN_vkEnumeratePhysicalDevices fn_enum_devs = (PFN_vkEnumeratePhysicalDevices)g_vkGetInstanceProcAddr(inst, "vkEnumeratePhysicalDevices");
    PFN_vkGetPhysicalDeviceProperties fn_get_props = (PFN_vkGetPhysicalDeviceProperties)g_vkGetInstanceProcAddr(inst, "vkGetPhysicalDeviceProperties");
    if (!fn_destroy || !fn_enum_devs || !fn_get_props) {
        if (fn_destroy) fn_destroy(inst, NULL);
        return 0;
    }

    uint32_t count = 0;
    if (fn_enum_devs(inst, &count, NULL) != VK_SUCCESS || count == 0) {
        fn_destroy(inst, NULL);
        return 0;
    }
    if (count > 16) {
        count = 16;
    }
    VkPhysicalDevice devs[16];
    if (fn_enum_devs(inst, &count, devs) != VK_SUCCESS || count == 0) {
        fn_destroy(inst, NULL);
        return 0;
    }

    int best_rank = 99;
    char best_name[256];
    best_name[0] = '\\0';
    for (uint32_t i = 0; i < count; i++) {
        VkPhysicalDeviceProperties props;
        memset(&props, 0, sizeof(props));
        fn_get_props(devs[i], &props);
        if (props.deviceType == VK_PHYSICAL_DEVICE_TYPE_CPU) {
            continue;
        }
        int rank = 5;
        if (props.deviceType == VK_PHYSICAL_DEVICE_TYPE_DISCRETE_GPU) rank = 1;
        else if (props.deviceType == VK_PHYSICAL_DEVICE_TYPE_INTEGRATED_GPU) rank = 2;
        else if (props.deviceType == VK_PHYSICAL_DEVICE_TYPE_VIRTUAL_GPU) rank = 3;
        if (rank < best_rank) {
            best_rank = rank;
            snprintf(best_name, sizeof(best_name), "%s", props.deviceName);
        }
    }

    fn_destroy(inst, NULL);
    if (best_rank == 99) {
        return 0;
    }
    if (name_out && name_cap > 0) {
        snprintf(name_out, name_cap, "%s", best_name);
    }
    return 1;
}
""")

special = {
    "vkGetInstanceProcAddr": """PFN_vkVoidFunction vkGetInstanceProcAddr(VkInstance instance, const char* pName) {
    pthread_once(&g_vk_once, init_vulkan_dl);
    if (g_vk_dummy_mode || !g_vkGetInstanceProcAddr || instance == (VkInstance)(uintptr_t)0x1) {
        return NULL;
    }
    return g_vkGetInstanceProcAddr(instance, pName);
}""",
    "vkGetDeviceProcAddr": """PFN_vkVoidFunction vkGetDeviceProcAddr(VkDevice device, const char* pName) {
    PFN_vkGetDeviceProcAddr fn = (PFN_vkGetDeviceProcAddr)resolve_vk_sym("vkGetDeviceProcAddr");
    if (g_vk_dummy_mode || !fn) {
        return NULL;
    }
    return fn(device, pName);
}""",
    "vkEnumerateInstanceVersion": """VkResult vkEnumerateInstanceVersion(uint32_t* pApiVersion) {
    PFN_vkEnumerateInstanceVersion fn = (PFN_vkEnumerateInstanceVersion)resolve_vk_sym("vkEnumerateInstanceVersion");
    if (!fn || g_vk_dummy_mode) {
        g_vk_dummy_mode = 1;
        if (pApiVersion) *pApiVersion = VK_MAKE_API_VERSION(0, 1, 2, 0);
        return VK_SUCCESS;
    }
    VkResult r = fn(pApiVersion);
    if (r != VK_SUCCESS || (pApiVersion && *pApiVersion < VK_MAKE_API_VERSION(0, 1, 2, 0))) {
        g_vk_dummy_mode = 1;
        if (pApiVersion) *pApiVersion = VK_MAKE_API_VERSION(0, 1, 2, 0);
        return VK_SUCCESS;
    }
    return r;
}""",
    "vkEnumerateInstanceExtensionProperties": """VkResult vkEnumerateInstanceExtensionProperties(const char* pLayerName, uint32_t* pPropertyCount, VkExtensionProperties* pProperties) {
    PFN_vkEnumerateInstanceExtensionProperties fn = (PFN_vkEnumerateInstanceExtensionProperties)resolve_vk_sym("vkEnumerateInstanceExtensionProperties");
    if (g_vk_dummy_mode || !fn) {
        if (pPropertyCount) *pPropertyCount = 0;
        return VK_SUCCESS;
    }
    return fn(pLayerName, pPropertyCount, pProperties);
}""",
    "vkEnumerateInstanceLayerProperties": """VkResult vkEnumerateInstanceLayerProperties(uint32_t* pPropertyCount, VkLayerProperties* pProperties) {
    PFN_vkEnumerateInstanceLayerProperties fn = (PFN_vkEnumerateInstanceLayerProperties)resolve_vk_sym("vkEnumerateInstanceLayerProperties");
    if (g_vk_dummy_mode || !fn) {
        if (pPropertyCount) *pPropertyCount = 0;
        return VK_SUCCESS;
    }
    return fn(pPropertyCount, pProperties);
}""",
    "vkCreateInstance": """VkResult vkCreateInstance(const VkInstanceCreateInfo* pCreateInfo, const VkAllocationCallbacks* pAllocator, VkInstance* pInstance) {
    PFN_vkCreateInstance fn = (PFN_vkCreateInstance)resolve_vk_sym("vkCreateInstance");
    if (!g_vk_dummy_mode && fn) {
        VkResult r = fn(pCreateInfo, pAllocator, pInstance);
        if (r == VK_SUCCESS && pInstance && *pInstance != VK_NULL_HANDLE) {
            g_last_instance = *pInstance;
            return r;
        }
    }
    g_vk_dummy_mode = 1;
    if (pInstance) *pInstance = (VkInstance)(uintptr_t)0x1;
    return VK_SUCCESS;
}""",
    "vkDestroyInstance": """void vkDestroyInstance(VkInstance instance, const VkAllocationCallbacks* pAllocator) {
    if (instance == VK_NULL_HANDLE || instance == (VkInstance)(uintptr_t)0x1) {
        return;
    }
    PFN_vkDestroyInstance fn = (PFN_vkDestroyInstance)resolve_vk_sym("vkDestroyInstance");
    if (fn) fn(instance, pAllocator);
    if (g_last_instance == instance) g_last_instance = VK_NULL_HANDLE;
}""",
    "vkEnumeratePhysicalDevices": """VkResult vkEnumeratePhysicalDevices(VkInstance instance, uint32_t* pPhysicalDeviceCount, VkPhysicalDevice* pPhysicalDevices) {
    if (g_vk_dummy_mode || instance == (VkInstance)(uintptr_t)0x1) {
        if (pPhysicalDeviceCount) *pPhysicalDeviceCount = 0;
        return VK_SUCCESS;
    }
    PFN_vkEnumeratePhysicalDevices fn = (PFN_vkEnumeratePhysicalDevices)resolve_vk_sym("vkEnumeratePhysicalDevices");
    if (!fn) {
        if (pPhysicalDeviceCount) *pPhysicalDeviceCount = 0;
        return VK_SUCCESS;
    }
    return fn(instance, pPhysicalDeviceCount, pPhysicalDevices);
}""",
}

for sym in needed:
    if sym in special:
        lines.append(special[sym])
        continue
    ret, raw_args = sigs[sym]
    arg_names = parse_arg_names(raw_args)
    call_args = ", ".join(arg_names)
    pfn_t = "PFN_" + sym
    if ret == "void":
        lines.append(f"""void {sym}({raw_args}) {{
    {pfn_t} fn = ({pfn_t})resolve_vk_sym("{sym}");
    if (fn) fn({call_args});
}}""")
    else:
        err_ret = "VK_ERROR_INITIALIZATION_FAILED" if ret == "VkResult" else f"({ret})0"
        lines.append(f"""{ret} {sym}({raw_args}) {{
    {pfn_t} fn = ({pfn_t})resolve_vk_sym("{sym}");
    if (!fn) return {err_ret};
    return fn({call_args});
}}""")

with open(out_c, "w", encoding="utf-8") as f:
    f.write("\n\n".join(lines) + "\n")
print(f"Generated {out_c} wrapping {len(needed)} Vulkan symbols.")
PYEOF

    gcc -O2 -fPIC -I"$VK_INC_DIR" -c "$VK_SHIM_C" -o "$VK_SHIM_O"

    mkdir -p "$STATIC_PREFIX/include" "$STATIC_PREFIX/lib"
    cp -a "$WHISPER_SRC/include/whisper.h" "$STATIC_PREFIX/include/"
    cp -a "$WHISPER_SRC/ggml/include"/*.h "$STATIC_PREFIX/include/"

    rm -f "$STATIC_PREFIX/lib/libwhisper_combined.a"
    ar -M <<EOF
CREATE $STATIC_PREFIX/lib/libwhisper_combined.a
ADDLIB $LIB_WHISPER_A
ADDLIB $LIB_GGML_A
ADDLIB $LIB_GGML_BASE_A
ADDLIB $LIB_GGML_CPU_A
ADDLIB $LIB_GGML_VK_A
ADDMOD $VK_SHIM_O
SAVE
END
EOF
    ranlib "$STATIC_PREFIX/lib/libwhisper_combined.a"
fi

cp -a "$STATIC_PREFIX/include/whisper.h" "$DIST_DIR/include/"
cp -a "$STATIC_PREFIX/include"/ggml*.h "$DIST_DIR/include/"
cp -a "$STATIC_PREFIX/lib/libwhisper_combined.a" "$DIST_DIR/lib/libwhisper.a"
ranlib "$DIST_DIR/lib/libwhisper.a"

# Verify every Vulkan/OpenMP symbol referenced by any object in libwhisper.a is resolved
# internally by vulkan_dlopen_shim.o (zero unresolved external vk* or gomp_* imports).
python3 - "$DIST_DIR/lib/libwhisper.a" <<'PYEOF'
import re, subprocess, sys
lib = sys.argv[1]
undef = set(re.findall(r'^\s+U\s+(\S+)$', subprocess.check_output(["nm", "-u", lib], text=True), re.M))
defn = set(re.findall(r'^[0-9a-f]+\s+[TWDBRCV]\s+(\S+)$', subprocess.check_output(["nm", "--defined-only", lib], text=True), re.M))
bad = sorted(s for s in (undef - defn) if re.match(r'^(vk[A-Z]|gomp_)', s))
if bad:
    print("ERROR: libwhisper.a unexpectedly has unresolved Vulkan or libgomp symbol imports:", file=sys.stderr)
    for s in bad:
        print(f"  U {s}", file=sys.stderr)
    sys.exit(1)
PYEOF

STDCPP_A="$(g++ -print-file-name=libstdc++.a)"
GCC_EH_A="$(g++ -print-file-name=libgcc_eh.a)"
cat > "$PKGCFG_DIR/whisper.pc" <<EOF
prefix=${DIST_DIR}
exec_prefix=${DIST_DIR}
libdir=${DIST_DIR}/lib
includedir=${DIST_DIR}/include

Name: Whisper
Description: Portable whisper.cpp static library (CPU + runtime dlopen Vulkan GPU)
Version: ${WHISPER_VER}
Libs: ${DIST_DIR}/lib/libwhisper.a ${STDCPP_A} ${GCC_EH_A} -ldl -lm -lpthread
Cflags: -I${DIST_DIR}/include -pthread
EOF

cp -a "$BUILD_CMAKE/output/bladeRF-remote" "$DIST_DIR/bin/bladerf-remote"
if command -v strip >/dev/null 2>&1; then
    strip --strip-unneeded "$DIST_DIR/bin/bladerf-remote"
fi
ln -sf bladerf-remote "$DIST_DIR/bin/bladeRF-remote"
ln -sf bladerf-remote "$DIST_DIR/bin/bladerf-tcp"

echo "Built portable static archives (libbladeRF.a, libusb-1.0.a, libportaudio.a, libwhisper.a) and bladerf-remote in $DIST_DIR"
