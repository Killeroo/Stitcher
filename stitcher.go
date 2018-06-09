// Alot of credit to gombine (https://github.com/r3s/gombine) for providing inspiration to some
// of the logic for obtaining image data (nioce code man)
// Basic image manipulation stuff (https://stackoverflow.com/a/35965499)

// We should be doing it like this
//https://github.com/r3s/gombine/blob/master/main.go
//https://golang.org/pkg/log/
//https://gobyexample.com/panic

package main

import (
	"image"
	"image/png"
	"image/draw"
	"os"
	"flag"
	"log"
)

type ImageData struct {
	img	   image.Image
	height int
	width  int
	path   string
}

// Gets data of an image object and returns info in an ImageData struct
func getImageData(img *image.Image, filename string) (ImageData, error) {
	data := new(ImageData)
	data.img = *img
	data.path = filename

	file, err := os.Open(filename)
	if err != nil {
		return *data, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return *data, err
	}

	data.height = config.Height
	data.width = config.Width

	return *data, nil
}

// TODO: Comparmentalise
func main() {
	log.SetPrefix("[STITCHER] ")
	log.SetFlags(0)
	flag.Parse()

	images := []*ImageData{}
	cols := 7 // TODO: Move to flag

	// Load data of all images
	// TODO: Add directory support
	// TODO: Check file is png
	for index, filename := range flag.Args() {

		// Try to open file
		imgFile, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer imgFile.Close()

		// Get image data (kept so data persists)
		imgData, _, err := image.Decode(imgFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load additional image data and store in array
		imgEntry, err := getImageData(&imgData, filename)
		if err != nil {
			log.Fatal(err)
		}
		images = append(images, &imgEntry)

		log.Printf("Loaded image [%d]: %s: x=%d y=%d", index+1, imgEntry.path, imgEntry.width, imgEntry.height)
	}

	// Check image sizes are the same
	// TODO: Allow image size flexibilty, switch to using average sizes
	maxWidth := images[0].width
	maxHeight := images[0].height
	for _, image := range images {
		if image.width > maxWidth || image.height > maxHeight {
			log.Fatal("Images are not the same size. Ensure images have the same dimensions then try again")
		}
	}

	// Work out new image dimensions
	var newImg ImageData
	var gridHeight int
	if len(images) % cols == 0 {
		newImg.height = maxHeight * len(images) / cols
		gridHeight = len(images) / cols
	} else {
		newImg.height = maxHeight * ((len(images) / cols)+1)
		gridHeight = (len(images) / cols)+1
	}
	newImg.width = maxWidth * cols
	log.Printf("New image dimensions (%dx%d grid): x=%d y=%d", cols, gridHeight, newImg.width, newImg.height)

	// Create new image
	r := image.Rectangle{image.Point{0,0}, image.Point{newImg.width, newImg.height}}
	rgba := image.NewRGBA(r)

	// Copy over image data from stored images
	var curHeight int
	var curWidth int
	for index, i := range images {
		draw.Draw(rgba, image.Rectangle{image.Point{0,0}, image.Point{curWidth,curHeight}}, i.img, image.Point{0,0}, draw.Src)
		log.Printf("Adding image [%d] @ %dx%d", index+1, curWidth, curHeight)
		if (index+1) % cols == 0 {
			curHeight += i.height
			curWidth = 0
		} else {
			curWidth += i.width
		}

	}

	// Save new image
	out, err := os.Create("./out.png")
	defer out.Close()
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(out, rgba)
	log.Println("New file created.")

	os.Exit(0)

}