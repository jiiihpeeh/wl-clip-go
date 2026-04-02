package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/jiiihpeeh/wl-clip-go/go/wlclip"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			if x < 50 && y < 50 {
				img.SetRGBA(x, y, color.RGBA{255, 0, 0, 255})
			} else if x >= 50 && y < 50 {
				img.SetRGBA(x, y, color.RGBA{0, 255, 0, 255})
			} else if x < 50 && y >= 50 {
				img.SetRGBA(x, y, color.RGBA{0, 0, 255, 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{255, 255, 0, 255})
			}
		}
	}

	if err := wlclip.SetImage(img); err != nil {
		fmt.Println("SetImage error:", err)
		return
	}
	fmt.Println("Set 100x100 quadrant image")

	bounds := img.Bounds()
	fmt.Printf("Original size: %dx%d\n", bounds.Dx(), bounds.Dy())

	retrieved, err := wlclip.GetImage()
	if err != nil {
		fmt.Println("GetImage error:", err)
		return
	}

	retrievedBounds := retrieved.Bounds()
	fmt.Printf("Retrieved size: %dx%d\n", retrievedBounds.Dx(), retrievedBounds.Dy())
}
