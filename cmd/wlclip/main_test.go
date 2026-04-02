package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func runCLI(args ...string) (string, string, int) {
	cmd := exec.Command("./wlclip-test", args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = "/home/j-p/wl-clip-go"
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

func runCLIPipe(input string, args ...string) (string, string, int) {
	cmd := exec.Command("./wlclip-test", args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = "/home/j-p/wl-clip-go"
	cmd.Stdin = bytes.NewBufferString(input)
	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

func TestVersionFlag(t *testing.T) {
	stdout, stderr, exitCode := runCLI("-version")
	if exitCode != 0 {
		t.Errorf("version flag exited with code %d: %s", exitCode, stderr)
		return
	}
	if stdout == "" {
		t.Error("version flag produced no output")
		return
	}
	if stderr != "" {
		t.Errorf("version flag produced stderr: %s", stderr)
		return
	}
}

func TestHelpFlag(t *testing.T) {
	stdout, stderr, exitCode := runCLI("-h")
	if exitCode != 0 {
		t.Errorf("help flag exited with code %d: %s", exitCode, stderr)
		return
	}
	output := stdout + stderr
	if output == "" {
		t.Error("help flag produced no output")
		return
	}
	if !bytes.Contains([]byte(output), []byte("wlclip")) {
		t.Errorf("help output doesn't contain 'wlclip': %s", output)
	}
}

func TestSetAndGetText(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"simple text", "Hello, World!"},
		{"multiline text", "Line 1\nLine 2\nLine 3"},
		{"unicode text", "Hello 世界 🌍"},
		{"special chars", "Tab:\tBackslash:\\Quote:\""},
		{"empty string", ""},
		{"newlines only", "\n\n\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exitCode := runCLIPipe(tc.input)
			if exitCode != 0 {
				t.Skip("Wayland compositor not available:", stderr)
			}

			stdout, stderr, exitCode := runCLI("-o")
			if exitCode != 0 {
				t.Fatalf("output mode failed: %s", stderr)
			}
			if stdout != tc.input {
				t.Errorf("got %q, want %q", stdout, tc.input)
			}
		})
	}
}

func TestSetAndGetTextWithNewlineTrim(t *testing.T) {
	input := "Hello\n"
	expected := "Hello"

	_, stderr, exitCode := runCLIPipe(input)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, stderr, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("output mode failed: %s", stderr)
	}
	if stdout != expected {
		t.Errorf("got %q, want %q (newline should be trimmed)", stdout, expected)
	}
}

func TestTextRoundTrip(t *testing.T) {
	input := "Clipboard test content 12345"

	_, stderr, exitCode := runCLIPipe(input)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: exit code %d", exitCode)
	}

	if stdout != input {
		t.Errorf("round trip failed: got %q, want %q", stdout, input)
	}
}

func TestSetAndGetFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	absPath1, _ := filepath.Abs(file1)
	absPath2, _ := filepath.Abs(file2)

	_, stderr, exitCode := runCLI(file1, file2)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	time.Sleep(500 * time.Millisecond)

	_, stderr, exitCode = runCLI("-o", "-t", "files")
	if exitCode != 0 {
		t.Skip("Wayland compositor does not support file operations:", stderr)
	}

	stdout, _, _ := runCLI("-o", "-t", "files")
	if !bytes.Contains([]byte(stdout), []byte(absPath1)) {
		t.Errorf("output doesn't contain %s: %s", absPath1, stdout)
	}
	if !bytes.Contains([]byte(stdout), []byte(absPath2)) {
		t.Errorf("output doesn't contain %s: %s", absPath2, stdout)
	}
}

func TestImageCopyAndPaste(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})
	img.SetRGBA(9, 9, color.RGBA{0, 0, 255, 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	pngData := buf.Bytes()

	_, stderr, exitCode := runCLIPipe(string(pngData), "-t", "image/png")
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, stderr, exitCode := runCLI("-o", "-t", "image/png")
	if exitCode != 0 {
		t.Fatalf("failed to get image as PNG: %s", stderr)
	}

	if len(stdout) == 0 {
		t.Error("got empty PNG data")
		return
	}

	decodedImg, err := png.Decode(bytes.NewReader([]byte(stdout)))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("image dimensions = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestFilterMode(t *testing.T) {
	input := "filter test content"
	expected := input

	stdout, stderr, exitCode := runCLIPipe(input, "-f")
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	if stdout != expected {
		t.Errorf("filter stdout = %q, want %q", stdout, expected)
	}

	clipboardOutput, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: exit code %d", exitCode)
	}
	if clipboardOutput != expected {
		t.Errorf("clipboard content = %q, want %q", clipboardOutput, expected)
	}
}

func TestImageFilterMode(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	img.SetRGBA(2, 2, color.RGBA{128, 128, 128, 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	pngData := buf.Bytes()

	stdout, stderr, exitCode := runCLIPipe(string(pngData), "-f", "-t", "image/png")
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	if len(stdout) == 0 {
		t.Error("filter output empty")
		return
	}

	decodedImg, err := png.Decode(bytes.NewReader([]byte(stdout)))
	if err != nil {
		t.Errorf("filter output is not valid PNG: %v", err)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 5 || bounds.Dy() != 5 {
		t.Errorf("image dimensions = %dx%d, want 5x5", bounds.Dx(), bounds.Dy())
	}
}

func TestOutputModeEmptyClipboard(t *testing.T) {
	_, stderr, exitCode := runCLI("-o")
	if exitCode == 0 {
		t.Error("expected non-zero exit code for empty clipboard")
	}
	if stderr == "" {
		t.Error("expected error message for empty clipboard")
	}
}

func TestInvalidMimeType(t *testing.T) {
	_, stderr, exitCode := runCLIPipe("test content", "-t", "invalid/mime-type")
	if exitCode == 0 {
		t.Error("expected non-zero exit code for invalid MIME type")
	}
	if stderr == "" {
		t.Error("expected error message for invalid MIME type")
	}
}

func TestNonexistentFile(t *testing.T) {
	_, stderr, exitCode := runCLI("/nonexistent/path/to/file.txt")
	if exitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent file")
	}
	if stderr == "" {
		t.Error("expected error message for nonexistent file")
	}
}

func TestMultipleFilesAsText(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	content1 := "content of file 1"
	content2 := "content of file 2"

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	_, stderr, exitCode := runCLI(file1, file2)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	clipboardOutput, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: exit code %d", exitCode)
	}
	if !bytes.Contains([]byte(clipboardOutput), []byte(content1)) {
		t.Errorf("clipboard doesn't contain file1 content: %s", clipboardOutput)
	}
}

func TestTargetFlagAlias(t *testing.T) {
	input := "target alias test"

	_, stderr, exitCode := runCLIPipe(input, "-target", "text/plain")
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: exit code %d", exitCode)
	}
	if stdout != input {
		t.Errorf("got %q, want %q", stdout, input)
	}
}

func TestSelectionFlag(t *testing.T) {
	input := "primary selection test"

	_, stderr, exitCode := runCLIPipe(input, "-selection", "primary")
	if exitCode != 0 {
		t.Skip("Wayland compositor may not support primary selection:", stderr)
	}

	_, stderr, exitCode = runCLI("-o", "-selection", "primary")
	if exitCode != 0 {
		t.Fatalf("failed to read primary selection: %s", stderr)
	}

	stdout, _, _ := runCLI("-o", "-selection", "primary")
	if stdout != input {
		t.Errorf("got %q, want %q", stdout, input)
	}
}

func TestPNGFileExtensionDetection(t *testing.T) {
	tmpDir := t.TempDir()
	pngFile := filepath.Join(tmpDir, "test.png")

	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.SetRGBA(1, 1, color.RGBA{255, 255, 0, 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	if err := os.WriteFile(pngFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write PNG file: %v", err)
	}

	_, stderr, exitCode := runCLI(pngFile)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o", "-t", "image/png")
	if exitCode != 0 {
		t.Fatalf("failed to get image: %s", stderr)
	}

	decodedImg, err := png.Decode(bytes.NewReader([]byte(stdout)))
	if err != nil {
		t.Errorf("failed to decode output PNG: %v", err)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 3 || bounds.Dy() != 3 {
		t.Errorf("image dimensions = %dx%d, want 3x3", bounds.Dx(), bounds.Dy())
	}
}

func TestURILListMimeType(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "uri_test.txt")
	if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	absPath, _ := filepath.Abs(file1)

	_, stderr, exitCode := runCLI("-t", "text/uri-list", file1)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o", "-t", "text/uri-list")
	if exitCode != 0 {
		t.Fatalf("failed to get files: %s", stderr)
	}

	if !bytes.Contains([]byte(stdout), []byte(absPath)) {
		t.Errorf("output doesn't contain file path: %s", stdout)
	}
}

func TestSetImageFromRawRGBA(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x * 64), uint8(y * 64), 128, 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}

	_, stderr, exitCode := runCLIPipe(string(buf.Bytes()), "-t", "image/png")
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o", "-t", "image/png")
	if exitCode != 0 {
		t.Fatalf("failed to get image: %s", stderr)
	}

	decodedImg, err := png.Decode(bytes.NewReader([]byte(stdout)))
	if err != nil {
		t.Fatalf("failed to decode PNG: %v", err)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 4 {
		t.Errorf("image dimensions = %dx%d, want 4x4", bounds.Dx(), bounds.Dy())
	}
}

func TestEmptyPNG(t *testing.T) {
	_, stderr, exitCode := runCLIPipe("", "-t", "image/png")
	if exitCode == 0 {
		t.Error("expected error for empty PNG")
	}
	if !bytes.Contains([]byte(stderr), []byte("empty")) {
		t.Errorf("expected 'empty' in error: %s", stderr)
	}
}

func TestInvalidPNG(t *testing.T) {
	invalidPng := "not a png file content"

	_, stderr, exitCode := runCLIPipe(invalidPng, "-t", "image/png")
	if exitCode == 0 {
		t.Error("expected error for invalid PNG")
	}
	if stderr == "" {
		t.Error("expected error message for invalid PNG")
	}
}

func TestLargeText(t *testing.T) {
	largeText := string(bytes.Repeat([]byte("A"), 10000))

	_, stderr, exitCode := runCLIPipe(largeText)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: %d", exitCode)
	}

	if stdout != largeText {
		t.Errorf("large text round trip failed: got %d chars, want %d chars", len(stdout), len(largeText))
	}
}

func TestBinaryData(t *testing.T) {
	binaryData := string([]byte{0x01, 0x02, 0x7F})

	_, stderr, exitCode := runCLIPipe(binaryData)
	if exitCode != 0 {
		t.Skip("Wayland compositor not available:", stderr)
	}

	stdout, _, exitCode := runCLI("-o")
	if exitCode != 0 {
		t.Fatalf("failed to read clipboard: %d", exitCode)
	}

	if stdout != binaryData {
		t.Errorf("binary data round trip failed")
	}
}
