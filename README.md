# go-bladerf

## build

Build the patched static `libbladeRF` (`libbladeRF.a` with static `libusb-1.0` netlink/usbfs + runtime RDMA `dlopen`), `bladerf-remote`, and `bladerf-mock` into `build/dist` with zero non-glibc shared library dependencies:

```sh
make
```

Package a relocatable multi-binary release tarball into `build/release/` (`linux/amd64` or `linux/arm64` via GitHub Actions [release workflow](./.github/workflows/release.yml)):

```sh
make release VERSION=v0.1.0
```

## mock server & testing

Start the standalone pure-Go mock `bladeRF 2.0 xA4` server (zero CGO/USB hardware needed):

```sh
./build/dist/bin/bladerf-mock -l 0.0.0.0:5555
```

`make test` automatically starts an embedded `mock.Server` when no physical bladeRF is connected so the entire test suite runs without hardware.

To run live RDMA (`librdmacm` `rsocket` over Linux SoftRoCE `rdma_rxe`) integration tests over an ephemeral SoftRoCE (`rdma_rxe`) `veth` link (automatically falling back to an ephemeral QEMU VM when the host kernel lacks `rdma_rxe`):

```sh
make test-rdma
```

## remote hardware over RDMA / TCP

Run `bladerf-remote` on the machine with the bladeRF hardware attached (e.g., Raspberry Pi):

```sh
./build/dist/bin/bladerf-remote -l 0.0.0.0:5555
```

- Uses RDMA (`librdmacm` / `rsocket` on `port + 1`, e.g., Linux `rdma_rxe` or `siw`) first when available, and automatically falls back to POSIX TCP on `port` (`5555`) when RDMA verbs/hardware are absent.
- Pass `--sim` to run without physical hardware for testing (`./build/dist/bin/bladerf-remote -l 127.0.0.1:5555 --sim`).

Connect from Go:

```go
// RDMA-first with automatic TCP fallback
dev, err := bladerf.OpenRDMA("raspberrypi.local:5555")
// Or force TCP
dev, err := bladerf.OpenTCP("raspberrypi.local:5555")
// Equivalent device identifier:
dev, err := bladerf.OpenWithDeviceIdentifier("rdma:addr=raspberrypi.local:5555")
```

Existing programs using `bladerf.Open()` or `bladerf.GetDeviceList()` can connect remotely without code changes by setting `BLADERF_REMOTE_ADDR`:

```sh
BLADERF_REMOTE_ADDR=raspberrypi.local:5555 ./your-bladerf-app
```

## tested on

`libbladeRF 2.6.1` (bundled patched build with RDMA/TCP backend) and `libbladeRF 2.2.1-MacPorts-20191220-45521019`
