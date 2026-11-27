.PHONY: all libbladerf bins release test test-rdma clean

VERSION ?= dev

all: libbladerf bins

libbladerf:
	./scripts/build_libbladerf.sh

bins: libbladerf
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o build/dist/bin/bladerf-mock ./cmd/bladerf-mock

release:
	./scripts/package_release.sh "$(VERSION)"

test: libbladerf
	go vet ./...
	go test -v ./...

test-rdma: all
	./scripts/test_rdma.sh

clean:
	rm -rf build
