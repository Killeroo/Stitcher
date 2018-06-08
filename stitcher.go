//https://stackoverflow.com/a/35965499

// We should be doing it like this
//https://github.com/r3s/gombine/blob/master/main.go
//https://golang.org/pkg/log/
//https://gobyexample.com/panic

package main

import (
	"fmt"
	"image"
	"image/png"
	"image/draw"
	"image/rgba"
	"os"
)

func main() {

	// Load images
	imgFile1, err := os.Open("test1.png")
	imgFile2, err := os.Open("test2.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	img1, _, err := image.Decode(imgFile1)
	img2, _, err := image.Decode(imgFile2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	// Find the starting position of the second image (Bottom left)
	sp2 := image.Point{img1.Bounds().Dx(), 0}

	// Create new rectangle for the second image
	r2 := image.Rectangle(sp2, sp2.Add(img2.Bounds()))

	// Now create a rectangle big enough for both images
	r := image.Rectangle{image.Point{0, 0}, r2.Max}

	// Create a new image
	rbga := image.newRGBA(r)

	// Now draw the two images into the new image
	draw.Draw(rgba, img1.Bounds(), img1, image.Point{0, 0}, draw.Src)
	draw.Draw(rgba, r2, img2, image.Point{0, 0}, draw.src)

	// NOTE: The height of the second image is taken into account
	// so if image 2 is taller than images 1 then extra height will be added to image 1
	// Also images 2 will be added to the right of image 1

	// Finally export image
	out, err := os.Create("./output.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	png.Encode(out, rgba)

	fmt.Println("New image created.\n")
}