# wl-clip-go: Next Session Documentation

## Project Overview

This project provides clipboard access for Wayland through Rust FFI bindings.

## Current Project Structure

```
wl-clip-go/
├── go/                          # Go library destination
│   └── wlclip/
│       └── clip.go              # <- TO BE CREATED
├── rust/                        # Rust FFI library
│   ├── src/lib.rs               # C-exported functions (already implemented)
│   ├── Cargo.toml               # staticlib, cdylib crate types
│   └── target/                  # Build output
│
├── internal/client/             # Existing Go client (socket-based daemon)
├── examples/                    # Example applications
├── cmd/                         # Command-line tools
└── go.mod                       # Go module definition
```

## Rust Library (rust/src/lib.rs)

Exported C functions:
- `wlclip_set_foreground(val bool)` - Set foreground mode
- `wlclip_get_text() -> WlClipString` - Get clipboard text
- `wlclip_set_text(text *c_char) -> WlClipInt` - Set clipboard text
- `wlclip_get_image() -> WlClipImage` - Get clipboard image (RGBA)
- `wlclip_set_image(rgba_data, len, width, height) -> WlClipInt` - Set clipboard image
- `wlclip_get_files() -> WlClipString` - Get file list (JSON array)
- `wlclip_set_files(json) -> WlClipInt` - Set file list from JSON
- `wlclip_free_string(ptr)` - Free string memory
- `wlclip_free_bytes(ptr, len)` - Free bytes memory

C Structs:
```c
struct WlClipString { char *ptr; size_t len; char *error; }
struct WlClipBytes  { uint8_t *ptr; size_t len; char *error; }
struct WlClipImage  { uint8_t *ptr; size_t len; uint32_t width, height; char *error; }
struct WlClipInt    { int32_t value; char *error; }
```

## Next Session Tasks

### 1. Create Go Library: `go/wlclip/clip.go`

Create CGO bindings that:
- Import the Rust compiled library (likely `libwlclip.a` or similar)
- Wrap the C functions in idiomatic Go
- Handle memory management (call free functions)
- Return proper Go types and errors

Key considerations:
- The Rust lib compiles to both `staticlib` and `cdylib` (Cargo.toml line 9)
- May need to build Rust library first: `cargo build --release` in rust/
- Library filename likely `librust wlclip.rlib` or `libwlclip.a`

### 2. Verify Library Builds

Steps:
1. Build Rust library: `cd rust && cargo build --release`
2. Create Go bindings that link to the compiled library
3. Run `go build ./go/...` to verify compilation

### 3. Run Tests

Create or run tests to verify:
- `GetText()` / `SetText()`
- `GetImage()` / `SetImage()`  
- `GetFiles()` / `SetFiles()`

## Build Commands Reference

```bash
# Build Rust library
cd rust
cargo build --release

# Verify Go compilation
go build ./go/...

# Run tests
go test ./go/...
```

## Dependencies

- Go 1.21+
- Rust toolchain
- CGO enabled (GCC/clang)

## Notes

- When running tests, ensure the daemon process is properly terminated after tests complete
- Tests may hang if Wayland clipboard operations block indefinitely - use `-timeout` flag with tests
- The Rust library may have issues with CGO bool type - use `char` in C declarations