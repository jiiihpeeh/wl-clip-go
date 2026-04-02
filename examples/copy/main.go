package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

var (
	readFromStdin = flag.Bool("stdin", false, "Read from stdin")
	readFile      = flag.String("file", "", "Read from file")
	showClipboard = flag.Bool("show", false, "Show current clipboard content")
)

func main() {
	flag.Parse()

	if *showClipboard {
		text, err := wlclip.GetText()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading clipboard: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(text)
		return
	}

	var data []byte
	var err error

	switch {
	case *readFromStdin:
		data, err = io.ReadAll(os.Stdin)
	case *readFile != "":
		data, err = os.ReadFile(*readFile)
	default:
		if flag.NArg() > 0 {
			data = []byte(flag.Arg(0))
		} else {
			fmt.Fprintf(os.Stderr, "Usage: copy [-stdin] [-file=<path>] [-show] [text]\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	format, err := wlclip.DetectImageFormat(data)
	if err == nil && format != "" {
		if err := wlclip.SetImageType(data, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Image (%s) copied to clipboard\n", format)
		return
	}

	if err := wlclip.SetText(string(data)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting clipboard: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Text copied to clipboard")
}
