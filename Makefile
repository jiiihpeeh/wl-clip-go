.PHONY: all build build-rust-libs build-rust build-go generate install clean test

RUST_DIR = rust
GO_DIR = go
LIBS_DIR = $(RUST_DIR)/libs
DESTDIR ?=
PREFIX ?= $(HOME)/.local
PROJECT := wlclip
BIN_DIR := $(DESTDIR)$(PREFIX)/$(PROJECT)

# Supported architectures for pre-built libs
SUPPORTED_ARCHS = x86_64-unknown-linux-gnu aarch64-unknown-linux-gnu

# Determine host architecture
HOST_ARCH := $(shell rustc -vV | grep ^host | cut -d' ' -f2)
CROSS_ARCHS = $(filter-out $(HOST_ARCH),$(SUPPORTED_ARCHS))

all: build

build: generate decompress build-go

generate:
	cd go/wlclip && go run gen.go

decompress:
	@go run internal/decompress/decompress.go

build-rust-libs:
	mkdir -p $(LIBS_DIR)
	for arch in $(SUPPORTED_ARCHS); do mkdir -p $(LIBS_DIR)/$$arch; done
	# Build native arch
	$(MAKE) build-rust-single TARGET=$(HOST_ARCH)
	# Cross-compile other archs (requires cross-compiler)
	@for arch in $(CROSS_ARCHS); do \
		$(MAKE) build-rust-single TARGET=$$arch || echo "Skipping $$arch (cross-compiler may not be installed)"; \
	done

build-rust-single:
	cd $(RUST_DIR) && cargo build --release --target $(TARGET)
	cp $(RUST_DIR)/target/$(TARGET)/release/libwlclip.a $(LIBS_DIR)/$(TARGET)/libwlclip.a

build-rust:
	cd $(RUST_DIR) && cargo build --release

build-go:
	cd . && go build -ldflags="-s -w" -o wlclip ./cmd/wlclip
	cd . && go build -ldflags="-s -w" -o wlclipd ./cmd/wlclipd
	@which upx > /dev/null && upx wlclipd || true

install: build
	install -Dm755 wlclip $(BIN_DIR)/wlclip
	@echo "Installed to $(BIN_DIR)/wlclip"
	@echo ""
	@echo "Add to PATH:"
	@echo "  For user install (~/.local/wlclip):"
	@echo "    echo 'export PATH=\"$$HOME/.local/wlclip:$$PATH\"' >> ~/.profile && . ~/.profile"
	@echo ""
	@echo "  For system install (/usr/wlclip):"
	@echo "    echo 'export PATH=/usr/wlclip:$$PATH' | sudo tee /etc/profile.d/wlclip.sh"

clean:
	cd $(RUST_DIR) && cargo clean
	cd $(GO_DIR) && go clean
	rm -rf $(LIBS_DIR)
	rm -f go/wlclip/clip_amd64.go go/wlclip/clip_arm64.go

test: build
	cd $(GO_DIR) && go test ./wlclip/...
