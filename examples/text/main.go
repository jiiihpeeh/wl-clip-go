package main

import (
	"fmt"
	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

func main() {
	if err := wlclip.SetText("Hello from wl-clip-go!"); err != nil {
		fmt.Println("SetText error:", err)
		return
	}

	text, err := wlclip.GetText()
	if err != nil {
		fmt.Println("GetText error:", err)
		return
	}

	fmt.Println("Clipboard text:", text)
}
