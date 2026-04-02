package wlclip

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"testing"
)

func TestSetAndGetText(t *testing.T) {
	SetForeground(false)

	testCases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"simple text", "Hello, World!"},
		{"multiline text", "Line 1\nLine 2\nLine 3"},
		{"unicode text", "Hello 世界 🌍"},
		{"special chars", "Tab:\tBackslash:\\Quote:\""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := SetText(tc.input); err != nil {
				t.Skip("Wayland compositor not available:", err)
			}

			got, err := GetText()
			if err != nil {
				t.Fatalf("GetText() error = %v", err)
			}
			if got != tc.input {
				t.Errorf("GetText() = %q, want %q", got, tc.input)
			}
		})
	}
}

func TestGetTextWithoutSet(t *testing.T) {
	_, err := GetText()
	if err != nil {
		t.Skip("Wayland compositor not available:", err)
	}
}

func TestSetFiles(t *testing.T) {
	SetForeground(false)

	testCases := []struct {
		name  string
		paths []string
	}{
		{"single file", []string{"/tmp/test.txt"}},
		{"multiple files", []string{"/tmp/file1.txt", "/tmp/file2.txt", "/tmp/file3.txt"}},
		{"with spaces", []string{"/tmp/file with spaces.txt"}},
		{"unix paths", []string{"/home/user/documents/file.pdf"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := SetFiles(tc.paths); err != nil {
				t.Skip("Wayland compositor not available:", err)
			}

			got, err := GetFiles()
			if err != nil {
				t.Fatalf("GetFiles() error = %v", err)
			}
			if len(got) != len(tc.paths) {
				t.Errorf("GetFiles() returned %d files, want %d", len(got), len(tc.paths))
			}
		})
	}
}

func TestSetFilesEmpty(t *testing.T) {
	err := SetFiles([]string{})
	if err == nil {
		t.Error("SetFiles([]string{}) should return error for empty paths")
	}
}

func TestSetFilesJSONFormat(t *testing.T) {
	SetForeground(false)

	paths := []string{"/tmp/a.txt", "/tmp/b.txt"}
	expectedJSON, _ := json.Marshal(paths)

	if err := SetFiles(paths); err != nil {
		t.Skip("Wayland compositor not available:", err)
	}

	got, err := GetFiles()
	if err != nil {
		t.Fatalf("GetFiles() error = %v", err)
	}

	gotJSON, _ := json.Marshal(got)
	if !bytes.Equal(gotJSON, expectedJSON) {
		t.Errorf("GetFiles() JSON = %q, want %q", string(gotJSON), string(expectedJSON))
	}
}

func TestSetAndGetImage(t *testing.T) {
	SetForeground(false)

	testCases := []struct {
		name   string
		width  int
		height int
		color  color.RGBA
	}{
		{"1x1 red", 1, 1, color.RGBA{255, 0, 0, 255}},
		{"10x10 blue", 10, 10, color.RGBA{0, 0, 255, 255}},
		{"2x3 green", 2, 3, color.RGBA{0, 255, 0, 255}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tc.width, tc.height))
			for y := 0; y < tc.height; y++ {
				for x := 0; x < tc.width; x++ {
					img.SetRGBA(x, y, tc.color)
				}
			}

			if err := SetImage(img); err != nil {
				t.Skip("Wayland compositor not available:", err)
			}

			got, err := GetImage()
			if err != nil {
				t.Fatalf("GetImage() error = %v", err)
			}

			bounds := got.Bounds()
			if bounds.Dx() != tc.width || bounds.Dy() != tc.height {
				t.Errorf("GetImage() size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tc.width, tc.height)
			}
		})
	}
}

func TestSetImageNonRGBA(t *testing.T) {
	SetForeground(false)

	bounds := image.Rect(0, 0, 5, 5)
	img := image.NewNRGBA64(bounds)

	if err := SetImage(img); err != nil {
		t.Skip("Wayland compositor not available:", err)
	}
}

func TestGetFilesWithoutSet(t *testing.T) {
	_, err := GetFiles()
	if err != nil {
		t.Skip("Wayland compositor not available:", err)
	}
}

func TestImageConversion(t *testing.T) {
	SetForeground(false)

	bounds := image.Rect(0, 0, 4, 4)
	img := image.NewRGBA(bounds)
	img.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})
	img.SetRGBA(3, 3, color.RGBA{0, 255, 0, 255})

	if err := SetImage(img); err != nil {
		t.Skip("Wayland compositor not available:", err)
	}

	retrieved, err := GetImage()
	if err != nil {
		t.Fatalf("GetImage() error = %v", err)
	}

	if retrieved.Bounds() != bounds {
		t.Errorf("GetImage() bounds = %v, want %v", retrieved.Bounds(), bounds)
	}
}
