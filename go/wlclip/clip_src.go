//go:build ignore

package wlclip

/*
#cgo LDFLAGS: ${SRCDIR}/../../rust/libs/ARCH_LIB_PATH/libwlclip.a -lwayland-client -lwayland-cursor -lrt -lm -ldl -lpthread
#cgo pkg-config: wayland-client

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* ptr;
    size_t len;
    char* error;
} WlClipString;

typedef struct {
    uint8_t* ptr;
    size_t len;
    char* error;
} WlClipBytes;

typedef struct {
    int32_t value;
    char* error;
} WlClipInt;

void wlclip_set_foreground(char val);
WlClipString wlclip_get_text();
WlClipInt wlclip_set_text(char* text);
WlClipBytes wlclip_get_image();
WlClipInt wlclip_set_image_type(uint8_t* image_data, size_t len, char* mime_type);
WlClipString wlclip_get_files();
WlClipInt wlclip_set_files(char* json);
void wlclip_free_string(char* ptr);
void wlclip_free_bytes(uint8_t* ptr, size_t len);
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"unsafe"
)

var (
	ErrClipboardEmpty   = errors.New("clipboard is empty or unavailable")
	ErrClipboardNoImage = errors.New("clipboard is empty or contains no image")
	ErrClipboardNoFiles = errors.New("clipboard is empty or contains no files")
	ErrInvalidImageData = errors.New("invalid image data")
	ErrEmptyImage       = errors.New("empty image data")
)

func SetForeground(blocking bool) {
	if blocking {
		C.wlclip_set_foreground(1)
	} else {
		C.wlclip_set_foreground(0)
	}
}

func GetText() (string, error) {
	result := C.wlclip_get_text()
	defer C.wlclip_free_string(result.ptr)

	if result.error != nil {
		return "", errors.New(C.GoString(result.error))
	}
	if result.ptr == nil {
		return "", ErrClipboardEmpty
	}
	return C.GoStringN(result.ptr, C.int(result.len)), nil
}

func SetText(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	result := C.wlclip_set_text(cText)
	if result.error != nil {
		err := C.GoString(result.error)
		C.free(unsafe.Pointer(result.error))
		return errors.New(err)
	}
	return nil
}

func GetImage() (image.Image, error) {
	result := C.wlclip_get_image()
	defer C.wlclip_free_bytes(result.ptr, C.size_t(result.len))

	if result.error != nil {
		return nil, errors.New(C.GoString(result.error))
	}
	if result.ptr == nil || result.len == 0 {
		return nil, ErrClipboardNoImage
	}

	sliceLen := int(result.len)
	cSlice := unsafe.Slice(result.ptr, sliceLen)
	imgData := make([]byte, sliceLen)
	for i := 0; i < sliceLen; i++ {
		imgData[i] = byte(cSlice[i])
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, ErrInvalidImageData
	}
	return img, nil
}

func SetImage(img image.Image) error {
	if img == nil {
		return ErrInvalidImageData
	}

	bounds := img.Bounds()
	if bounds.Empty() {
		return ErrEmptyImage
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ErrInvalidImageData
	}

	pngData := buf.Bytes()
	cMimeType := C.CString("image/png")
	defer C.free(unsafe.Pointer(cMimeType))

	cData := (*C.uint8_t)(unsafe.Pointer(&pngData[0]))
	result := C.wlclip_set_image_type(cData, C.size_t(len(pngData)), cMimeType)

	if result.error != nil {
		err := C.GoString(result.error)
		C.free(unsafe.Pointer(result.error))
		return errors.New(err)
	}
	return nil
}

func SetImageType(imageData []byte, mimeType string) error {
	if len(imageData) == 0 {
		return ErrEmptyImage
	}

	cMimeType := C.CString(mimeType)
	defer C.free(unsafe.Pointer(cMimeType))

	cData := (*C.uint8_t)(unsafe.Pointer(&imageData[0]))
	result := C.wlclip_set_image_type(cData, C.size_t(len(imageData)), cMimeType)

	if result.error != nil {
		err := C.GoString(result.error)
		C.free(unsafe.Pointer(result.error))
		return errors.New(err)
	}
	return nil
}

func DetectImageFormat(data []byte) (string, error) {
	if len(data) < 4 {
		return "", ErrInvalidImageData
	}

	switch {
	case isPNG(data):
		return "image/png", nil
	case isJPEG(data):
		return "image/jpeg", nil
	case isGIF(data):
		return "image/gif", nil
	case isWebP(data):
		return "image/webp", nil
	case isTIFF(data):
		return "image/tiff", nil
	case isBMP(data):
		return "image/bmp", nil
	case isAVIF(data):
		return "image/avif", nil
	case isJXL(data):
		return "image/jxl", nil
	case isPDF(data):
		return "application/pdf", nil
	}
	return "", ErrInvalidImageData
}

func isPNG(data []byte) bool {
	return data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47
}

func isJPEG(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF
}

func isGIF(data []byte) bool {
	return len(data) >= 6 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && (data[3] == 0x38 && (data[4] == 0x37 || data[4] == 0x39) && data[5] == 0x61)
}

func isWebP(data []byte) bool {
	return len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 && data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50
}

func isTIFF(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00) ||
		(data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A)
}

func isBMP(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D
}

func isPDF(data []byte) bool {
	return len(data) >= 4 && data[0] == 0x25 && data[1] == 0x50 && data[2] == 0x44 && data[3] == 0x46
}

func isAVIF(data []byte) bool {
	return len(data) >= 12 && data[4] == 0x66 && data[5] == 0x74 && data[6] == 0x79 && data[7] == 0x70 && data[8] == 0x61 && data[9] == 0x76 && data[10] == 0x69 && data[11] == 0x66
}

func isJXL(data []byte) bool {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0x0A {
		return true
	}
	return len(data) >= 12 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x0C && data[4] == 0x4A && data[5] == 0x58 && data[6] == 0x4C && data[7] == 0x20
}

func GetFiles() ([]string, error) {
	result := C.wlclip_get_files()
	defer C.wlclip_free_string(result.ptr)

	if result.error != nil {
		return nil, errors.New(C.GoString(result.error))
	}
	if result.ptr == nil {
		return nil, ErrClipboardNoFiles
	}

	jsonStr := C.GoStringN(result.ptr, C.int(result.len))
	var files []string
	if err := json.Unmarshal([]byte(jsonStr), &files); err != nil {
		return nil, err
	}
	return files, nil
}

func SetFiles(paths []string) error {
	if len(paths) == 0 {
		return errors.New("empty file list")
	}

	jsonData, err := json.Marshal(paths)
	if err != nil {
		return err
	}

	cJSON := C.CString(string(jsonData))
	defer C.free(unsafe.Pointer(cJSON))

	result := C.wlclip_set_files(cJSON)
	if result.error != nil {
		err := C.GoString(result.error)
		C.free(unsafe.Pointer(result.error))
		return errors.New(err)
	}
	return nil
}
