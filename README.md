# wl-clip-go

A **Go library** for Wayland clipboard access via Rust FFI. No external binaries needed - just import and use in your Go project.

## Why?

No more need for CLI tools like `wl-paste`/`wl-copy` installed on the system - embed clipboard access directly in your Go app.

## Credits

Built on top of [wl-clipboard-rs](https://github.com/YaLTeR/wl-clipboard-rs) by YaLTeR.

## Features

- **Text clipboard**: Copy/paste text with proper UTF-8 handling
- **Image clipboard**: Copy/paste images with format preservation (PNG, JPEG, GIF, WebP, TIFF, BMP, AVIF, JXL, PDF)
- **File clipboard**: Copy files via `text/uri-list` for file managers
- **Content auto-detection**: Automatically detects image format via magic bytes
- **Selection support**: clipboard, primary, or secondary selection
- **Daemon mode**: Automatic daemon startup for reliable operation
- **Multi-arch**: Pre-built static libraries for `amd64` and `arm64`

## Requirements

- **Go** 1.21+
- **Wayland compositor** (tested with KDE)
- **Linux** (Wayland only)

No Rust or cross-compilers needed - pre-built static libraries are shipped.

## Quick Start

```go
import "github.com/jiiihpeeh/wl-clip-go/go/wlclip"

func main() {
    // Copy text to clipboard
    wlclip.SetText("hello")

    // Read clipboard
    text, _ := wlclip.GetText()
    fmt.Println(text)
}
```

```bash
go build -o myapp .
```

Pre-built static libraries for `amd64` and `arm64` are included - no Rust toolchain needed.

## Library Usage

```go
import "github.com/jiiihpeeh/wl-clip-go/go/wlclip"
```

### Text operations

```go
// Copy text to clipboard
err := wlclip.SetText("hello world")

// Get text from clipboard
text, err := wlclip.GetText()
```

### Image operations

```go
// Set image from image data (PNG, JPEG, GIF, WebP, TIFF, BMP, AVIF, JXL)
// Format is automatically detected via magic bytes
err := wlclip.SetImage(imageBytes)

// Set image with specific MIME type
err := wlclip.SetImageType(imageBytes, "image/jpeg")

// Get image - returns whatever format is in clipboard
img, err := wlclip.GetImage()
```

### File operations

```go
// Copy files to clipboard (for file managers)
err := wlclip.SetFiles([]string{"/path/to/file1.txt", "/path/to/file2.txt"})

// Get files from clipboard
files, err := wlclip.GetFiles()
```

### Foreground mode

```go
// Control whether operations block until clipboard is consumed
// Default is false (returns immediately)
wlclip.SetForeground(false) // returns immediately
wlclip.SetForeground(true)  // blocks (more reliable)
```

## CLI Binaries (Optional)

If you also want the command-line tools:

```bash
# Build from source (no Rust needed!)
make

# Install to ~/.local/wlclip/
make install

# Or system-wide install
sudo make install PREFIX=/usr
```

After installation, add to PATH:

```bash
# For user install (~/.local/wlclip)
echo 'export PATH="$HOME/.local/wlclip:$PATH"' >> ~/.profile && . ~/.profile

# For system install
sudo tee /etc/profile.d/wlclip.sh << 'EOF'
export PATH=/usr/wlclip:$PATH
EOF
```

## CLI Usage

### CLI Options

```
  -selection <name>   Selection: clipboard, primary, secondary (default: clipboard)
  -i                  Input mode: copy stdin to clipboard (default when piping)
  -o                  Output mode: print clipboard to stdout
  -t, -target <type>  MIME type (e.g., image/png, text/plain, text/uri-list)
  -f                  Filter mode: copy stdin to stdout AND clipboard
  -clear              Clear the clipboard contents
  -h, -help           Show help
  -version            Show version
```

### Copy text to clipboard

```bash
echo "hello" | wlclip
wlclip < file.txt
```

### Output clipboard contents

```bash
wlclip -o
```

### Copy image to clipboard

```bash
# Auto-detects format via magic bytes (PNG, JPEG, GIF, WebP, TIFF, BMP, AVIF, JXL, PDF)
wlclip < image.png
wlclip image.jpg
wlclip image.webp

# Explicit MIME type
wlclip -t image/png < image.png
```

### Output clipboard as image

```bash
wlclip -o
wlclip -o -t image/png
```

### Copy files to clipboard

```bash
wlclip file1.txt file2.txt
```

### Output clipboard as file list

```bash
wlclip -o -t text/uri-list
```

### Filter mode (pipe through clipboard)

```bash
cat file.txt | wlclip -f
echo "test" | wlclip -f | sed 's/a/b/'
```

### Clear clipboard

```bash
wlclip -clear
```

## Architecture

```
wlclip (CLI)
    └── client (IPC via Unix socket)
            └── wlclipd (daemon)
                    └── Rust FFI (wl-clipboard-rs)
                            └── Wayland compositor
```

### Static Linking

Pre-built static libraries for `amd64` and `arm64` are shipped in `rust/libs/`. The Go build automatically selects the correct library based on `GOARCH`.

- `rust/libs/x86_64-unknown-linux-gnu/libwlclip.a`
- `rust/libs/aarch64-unknown-linux-gnu/libwlclip.a`

**Note**: The Rust library itself is statically linked, but Wayland client libraries (`libwayland-client`, `libwayland-cursor`) are dynamically linked. These must be installed on the target system:

```bash
# Ubuntu/Debian
sudo apt install libwayland-dev

# Fedora
sudo dnf install wayland-devel

# Arch Linux
sudo pacman -S wayland
```

The build uses `pkg-config` to find Wayland libraries automatically.

### Rebuilding the Libraries

If you need to rebuild the static libraries (e.g., after changing Rust code):

```bash
# Install cross-compiler for arm64
sudo apt install gcc-aarch64-linux-gnu

# Rebuild libraries for both architectures
make build-rust-libs
```

## Build

```bash
# Build both (uses pre-built static libs)
make

# Or manually build Go binaries
go build -o wlclip ./cmd/wlclip
go build -o wlclipd ./cmd/wlclipd

# Run tests
make test

# Clean build artifacts
make clean

# Install
make install
```

## Examples

See `examples/` directory for usage examples:
- `text/` - Text copy/paste
- `image/` - Image copy/paste (RGBA bytes)
- `png/` - PNG encoding/decoding
- `files/` - File clipboard operations
- `copy/` - Copy functionality

## Status

This project is functional but **not extensively tested**. It has been tested on a limited set of Wayland compositors (KDE) and configurations. More testing across different compositors, distributions, and use cases is needed before production use.

## Known Issues

### foreground(false) race condition

With some compositors (notably KDE Plasma), setting `foreground(false)` may cause clipboard data to not be immediately available for paste. The daemon mode mitigates this.

**Workaround**: Use the default `foreground(false)` mode (returns immediately) or ensure a clipboard consumer is available.

### File clipboard limitations

Some compositors don't properly support `text/uri-list` for file operations. The `TestSetAndGetFiles` test may be skipped on such systems.

### Empty clipboard detection

When using auto-detect output mode on an empty clipboard, behavior depends on the compositor. Some compositors may output empty text while others may produce an error.

## License

MIT License - see [LICENSE](LICENSE)
