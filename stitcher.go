// stitcher.go - Matthew Carney [matthewcarney64@gmail.com
// Alot of credit to gombine (https://github.com/r3s/gombine) for providing inspiration to some
// of the logic for obtaining image data (nioce code man)
// Basic image manipulation stuff (https://stackoverflow.com/a/35965499)

package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const iconText = `
       (o\---/o)
        ( . . )
       _( (T) )_
      / /     \ \
     / /       \ \
     \_)       (_/
       \   _   /
       _)  |  (_
      (___,'.___)   [STITCHER] - SPRITESHEET GENERATOR VERSION 1.0
`

type ImageData struct {
	img    image.Image
	height int
	width  int
	path   string
}

var fileCount int
var images []*ImageData

// Creates a new png using an array of image data. First function checks that all images are the same
// size, then works out the dimensions of the new image, after which a new image is created and each image
// in the list is copied over to the new image then saved.
func saveNewImage(images []*ImageData, cols int, outFile string) error {
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
	if len(images)%cols == 0 {
		newImg.height = maxHeight * len(images) / cols
		gridHeight = len(images) / cols
	} else {
		newImg.height = maxHeight * ((len(images) / cols) + 1)
		gridHeight = (len(images) / cols) + 1
	}
	newImg.width = maxWidth * cols
	log.Printf("New image dimensions (%dx%d grid): x=%d y=%d", cols, gridHeight, newImg.width, newImg.height)

	// Create new image
	r := image.Rectangle{image.Point{0, 0}, image.Point{newImg.width, newImg.height}}
	rgba := image.NewRGBA(r)

	// Copy over image data from stored images
	var curHeight int
	var curWidth int
	for index, i := range images {
		draw.Draw(rgba, image.Rectangle{image.Point{curWidth, curHeight}, image.Point{curWidth + i.width, curHeight + i.height}}, i.img, image.Point{0, 0}, draw.Src)
		log.Printf("Adding image [%d] @ %dx%d", index+1, curWidth, curHeight)
		if (index+1)%cols == 0 {
			curHeight += i.height
			curWidth = 0
		} else {
			curWidth += i.width
		}
	}

	// Save new image
	log.Println("Saving new file...")
	out, err := os.Create("./" + outFile + ".png")
	defer out.Close()
	if err != nil {
		return err
	}
	png.Encode(out, rgba)
	log.Println(outFile + ".png created.")

	return nil
}

// Gets data of an image object and returns info in an ImageData struct
// Gets the config data of an image (height, width and path) and stores the
// result in an ImageData struct which is then returned
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

// Checks if the file is a PNG.  First function opens the file, then 'sniffs' the first 512 bytes,
// these bytes are then given to the detectcontenttype function (hacky i know) which ascertains
// the file type, if it is a PNG then true is returned, else false is returned
func isPNG(path string) (bool, error) {

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}

	// 'Sniff' first 512 bytes to determine type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false, err
	}

	// Reset read pointer
	file.Seek(0, 0)

	fileType := http.DetectContentType(buffer)
	if strings.Compare(fileType, "image/png") == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func handleDirectory(arg string) {
	// Read all files in directory
	files, err := ioutil.ReadDir(arg)
	if err != nil {
		log.Fatal(err)
	}

	// Load data of all png we find
	for _, file := range files {
		if !file.Mode().IsDir() {
			filePath := path.Join(arg, file.Name())
			handleIndividualFile(filePath)
		}
	}
}

func handleIndividualFile(arg string) {
	// Check file type
	if f, _ := isPNG(arg); f == true { // Handle PNG

		// Try to open file
		imgFile, err := os.Open(arg)
		if err != nil {
			log.Fatal(err)
		}
		defer imgFile.Close()

		// Get image object data
		imgData, _, err := image.Decode(imgFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load additional image data and store in array
		imgEntry, err := getImageData(&imgData, arg)
		if err != nil {
			log.Fatal(err)
		}
		images = append(images, &imgEntry)

		fileCount++
		log.Printf("Loaded image [%d]: %s: %dx%d", fileCount, imgEntry.path, imgEntry.width, imgEntry.height)

	} else { // Handle GIF
		// File not supported
	}

}

func usage() {
	fmt.Println("stitcher [options] <file1> <file2> <dir1> ...")
	fmt.Println("\nOptions")
	flag.PrintDefaults()
	fmt.Println("\nExample: stitcher -cols 5 -name myfile C:\\Images C:\\SpecificImage\\test.png")
	os.Exit(0)
}

func main() {
	// Setup flags, log and icon
	fmt.Println(iconText)
	log.SetPrefix("[STITCHER] ")
	log.SetFlags(0)
	cols := flag.Int("cols", 4, "Number of columns to use on generated spritesheet")
	filename := flag.String("name", "out", "Name of generated spritesheet")
	flag.Usage = usage
	flag.Parse()

	// Bail out if no additional arguments
	if len(flag.Args()) == 0 {
		log.Fatal("No files or folder provided")
	}

	// Go through arguments and load image data
	for _, arg := range flag.Args() {

		// Get file type (file or dir)
		fi, err := os.Stat(arg)
		if err != nil {
			log.Fatal(err)
		}

		// Work out if path in argument points to a directory or a file
		if fi.Mode().IsDir() {
			handleDirectory(arg)
		} else {
			handleIndividualFile(arg)
		}
	}

	// Save our new image
	err := saveNewImage(images, *cols, *filename)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)

}
