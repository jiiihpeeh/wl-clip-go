# Clip - Wayland Clipboard Access

## Code Generation

This package uses arch-specific cgo bindings generated from a template.

- `clip_src.go` - Source template with placeholders
- `gen.go` - Generator that produces `clip_amd64.go` and `clip_arm64.go`
- `clip_amd64.go` - Generated for amd64 (gitignored)
- `clip_arm64.go` - Generated for arm64 (gitignored)

To regenerate arch-specific files after editing `clip_src.go`:

```bash
cd go/wlclip && go run gen.go
```

Or use `make generate` from the repo root.

## Building

The library requires Wayland development libraries:

```bash
# Ubuntu/Debian
sudo apt install libwayland-dev

# Fedora
sudo dnf install wayland-devel

# Arch Linux
sudo pacman -S wayland
```

Pre-built static libraries for `amd64` and `arm64` are included in `rust/libs/`. No Rust toolchain needed to build.
