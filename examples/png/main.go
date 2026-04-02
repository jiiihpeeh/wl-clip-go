package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 128, 255, 255})
		}
	}

	if err := wlclip.SetImage(img); err != nil {
		fmt.Println("SetImage error:", err)
		return
	}
	fmt.Println("Set image")

	retrieved, err := wlclip.GetImage()
	if err != nil {
		fmt.Println("GetImage error:", err)
		return
	}
	fmt.Printf("Retrieved image: %dx%d\n", retrieved.Bounds().Dx(), retrieved.Bounds().Dy())

	pngFile, err := os.Create("/tmp/test_image.png")
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer pngFile.Close()

	if err := png.Encode(pngFile, retrieved); err != nil {
		fmt.Println("Failed to encode PNG:", err)
		return
	}
	fmt.Println("Saved to /tmp/test_image.png")
}
