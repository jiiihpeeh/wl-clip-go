package main

import (
	"flag"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/jiiihpeeh/wl-clip-go/internal/client"
)

var (
	selection   = flag.String("selection", "clipboard", "Selection to use: clipboard, primary, secondary")
	outMode     = flag.Bool("o", false, "Output clipboard contents")
	inMode      = flag.Bool("i", false, "Input to clipboard (default if neither -i nor -o)")
	mimeType    = flag.String("t", "", "MIME type (default: autodetect)")
	filterMode  = flag.Bool("f", false, "Filter mode: copy stdin to stdout")
	target      = flag.String("target", "", "Alias for -t")
	helpFlag    = flag.Bool("h", false, "Show help")
	versionFlag = flag.Bool("version", false, "Show version")
	clearFlag   = flag.Bool("clear", false, "Clear clipboard contents")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `wlclip - Wayland clipboard utility

Usage: wlclip [options] [file...]

Options:
  -selection <name>   Selection: clipboard, primary, secondary (default: clipboard)
  -i                  Input mode: copy stdin to clipboard (default)
  -o                  Output mode: print clipboard to stdout
  -t, -target <type>  MIME type (e.g., image/png, text/plain)
  -f                  Filter mode: copy stdin to stdout (and clipboard if -i)
  -clear              Clear the clipboard
  -h, -help           Show this help
  -version            Show version

With no options or files, reads from stdin and copies to clipboard.
With -o or without files and stdin is a tty, outputs clipboard contents.

Examples:
  echo "hello" | wlclip                    Copy text to clipboard
  wlclip -o                               Print clipboard contents
  wlclip -o -t image/png                   Output clipboard as PNG
  wlclip -t image/png < image.png         Copy image to clipboard
  wlclip file1.txt file2.txt              Copy files to clipboard
  wlclip -f < file.txt                    Filter: copy stdin to stdout and clipboard
`)
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return
	}

	if *versionFlag {
		fmt.Println("wlclip 0.1.0")
		return
	}

	c, err := client.New()
	if err != nil {
		if err := client.EnsureDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		c, err = client.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: failed to connect to daemon: %v\n", err)
			os.Exit(1)
		}
	}
	defer c.Close()

	args := flag.Args()

	if *outMode {
		outputClipboard(c)
		return
	}

	if *filterMode {
		filter(c)
		return
	}

	if *clearFlag {
		if err := c.SetText(""); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) > 0 {
		handleFiles(c, args)
		return
	}

	if !isTerminal(os.Stdin) {
		handleStdin(c)
		return
	}

	outputClipboard(c)
}

func outputClipboard(c *client.Client) {
	mt := getMimeType()

	if mt == "image/png" || mt == "" {
		pngData, err := c.GetImage()
		if err == nil && len(pngData) > 0 {
			os.Stdout.Write(pngData)
			return
		}
	}

	if mt == "text/uri-list" || mt == "files" || mt == "" {
		files, err := c.GetFiles()
		if err == nil && len(files) > 0 {
			for _, f := range files {
				fmt.Println(f)
			}
			return
		}
	}

	text, err := c.GetText()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
		os.Exit(1)
	}
	if text == "" {
		fmt.Fprintf(os.Stderr, "wlclip: clipboard is empty\n")
		os.Exit(1)
	}
	fmt.Print(text)
}

func handleStdin(c *client.Client) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: reading stdin: %v\n", err)
		os.Exit(1)
	}

	mt := getMimeType()
	detectedMt := detectImageFormat(data)

	if mt == "image/png" || detectedMt == "image/png" || isPNG(data) {
		if err := c.SetImage(data); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if detectedMt != "" && mt == "" {
		mt = detectedMt
	}

	if mt != "" && !isValidMimeType(mt) {
		fmt.Fprintf(os.Stderr, "wlclip: invalid MIME type: %s\n", mt)
		os.Exit(1)
	}

	if detectedMt != "" && (strings.HasPrefix(detectedMt, "image/") || detectedMt == "application/pdf") {
		if err := c.SetImageType(data, detectedMt); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	text := string(data)
	if len(text) > 0 && strings.TrimRight(text, "\r\n") != "" {
		if strings.HasSuffix(text, "\r\n") {
			text = text[:len(text)-2]
		} else if strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\r") {
			text = text[:len(text)-1]
		}
	}
	if err := c.SetText(text); err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
		os.Exit(1)
	}
}

func handleFiles(c *client.Client, paths []string) {
	mt := getMimeType()

	data, err := os.ReadFile(paths[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
		os.Exit(1)
	}

	detectedMt := detectImageFormat(data)

	if mt == "image/png" || detectedMt == "image/png" || isPNG(data) {
		if err := c.SetImage(data); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if mt == "text/uri-list" || mt == "files" {
		var filePaths []string
		for _, p := range paths {
			absPath, err := filepath.Abs(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
				os.Exit(1)
			}
			filePaths = append(filePaths, absPath)
		}
		if err := c.SetFiles(filePaths); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if detectedMt != "" && (strings.HasPrefix(detectedMt, "image/") || detectedMt == "application/pdf") {
		if err := c.SetImageType(data, detectedMt); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := c.SetText(string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
		os.Exit(1)
	}
}

func filter(c *client.Client) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: reading stdin: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(data)

	mt := getMimeType()
	detectedMt := detectImageFormat(data)

	if mt == "image/png" || detectedMt == "image/png" || isPNG(data) {
		if err := c.SetImage(data); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if detectedMt != "" && (strings.HasPrefix(detectedMt, "image/") || detectedMt == "application/pdf") {
		if err := c.SetImageType(data, detectedMt); err != nil {
			fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := c.SetText(string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "wlclip: %v\n", err)
		os.Exit(1)
	}
}

func getMimeType() string {
	if *mimeType != "" {
		return *mimeType
	}
	if *target != "" {
		return *target
	}
	return ""
}

func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	switch {
	case isPNG(data):
		return "image/png"
	case isJPEG(data):
		return "image/jpeg"
	case isGIF(data):
		return "image/gif"
	case isWebP(data):
		return "image/webp"
	case isTIFF(data):
		return "image/tiff"
	case isBMP(data):
		return "image/bmp"
	case isPDF(data):
		return "application/pdf"
	case isAVIF(data):
		return "image/avif"
	case isJXL(data):
		return "image/jxl"
	}
	return ""
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

func imageExtensions() map[string]bool {
	return map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".webp": true,
		".tiff": true,
		".tif":  true,
		".bmp":  true,
		".pdf":  true,
		".avif": true,
		".jxl":  true,
	}
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := imageExtensions()[ext]
	return ok
}

var validMimeTypes = map[string]bool{
	"text/plain":               true,
	"text/plain;charset=utf-8": true,
	"text/html":                true,
	"text/css":                 true,
	"text/csv":                 true,
	"text/javascript":          true,
	"application/json":         true,
	"application/xml":          true,
	"application/octet-stream": true,
	"image/png":                true,
	"image/jpeg":               true,
	"image/gif":                true,
	"image/webp":               true,
	"image/tiff":               true,
	"image/bmp":                true,
	"image/avif":               true,
	"image/jxl":                true,
	"application/pdf":          true,
	"image/svg+xml":            true,
	"text/uri-list":            true,
}

func isValidMimeType(mt string) bool {
	if _, ok := validMimeTypes[mt]; ok {
		return true
	}
	_, _, err := mime.ParseMediaType(mt)
	if err != nil {
		return false
	}
	mtLower := strings.ToLower(mt)
	return strings.HasPrefix(mtLower, "text/") || strings.HasPrefix(mtLower, "image/")
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
