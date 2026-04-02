package main

import (
	"fmt"

	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

func main() {
	files := []string{
		"/tmp/test_file1.txt",
		"/tmp/test_file2.txt",
		"/tmp/test_file3.txt",
	}

	if err := wlclip.SetFiles(files); err != nil {
		fmt.Println("SetFiles error:", err)
		return
	}
	fmt.Println("Set files:", files)

	retrieved, err := wlclip.GetFiles()
	if err != nil {
		fmt.Println("GetFiles error:", err)
		return
	}

	fmt.Println("Retrieved files:", retrieved)
	fmt.Printf("Count: %d files\n", len(retrieved))
}
