// stitcher.go - Matthew Carney [matthewcarney64@gmail.com
// Alot of credit to gombine (https://github.com/r3s/gombine) for providing inspiration to some
// of the logic for obtaining image data (nioce code man)
// Basic image manipulation stuff (https://stackoverflow.com/a/35965499)

package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
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
      (___,'.___)   [STITCHER] - SPRITESHEET GENERATOR VERSION 1.1
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
	if len(images) == 0 {
		log.Fatal("No images have been loaded, cannot stitch images together. Check images and try again.")
	}

	// TODO: Allow image size flexibilty, switch to using average sizes
	// Check image sizes are the same
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

// Laods the image data from a PNG and returns it in an ImageData struct
func extractPNGData(filename string, imgData image.Image) (ImageData, error) {
	data := new(ImageData)

	// Try to open file
	file, err := os.Open(filename)
	if err != nil {
		return *data, err
	}
	defer file.Close()

	// Get image properties
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return *data, err
	}

	// Load all the data into the struct
	data.img = imgData
	data.path = filename
	data.height = config.Height
	data.width = config.Width

	return *data, nil
}

func extractGIFData(filename string) {
	log.Printf("Attempting to load image data from a GIF [%s]\n", filename)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	r := bufio.NewReader(file)
	g, err := gif.DecodeAll(r)
	if err != nil {
		log.Fatal(err)
	}

	for _, image := range g.Image {
		data := new(ImageData)
		data.img = image
		data.path = filename
		data.height = g.Config.Height
		data.width = g.Config.Width

		fileCount++
		log.Printf("Loaded image [%d]: %s: %dx%d", fileCount, data.path, data.width, data.height)

		images = append(images, data)
	}
}

// Checks the file at the path is an image file with the specified extension
func isImageFile(path string, ext string) bool {
	// Try opening the file
	file, err := os.Open(path)
	if err != nil {
		return false
	}

	// 'Sniff' first 512 bytes to determine type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}

	// Reset read pointer
	file.Seek(0, 0)

	// Detect that file type
	fileType := http.DetectContentType(buffer)
	if strings.Compare(fileType, "image/"+ext) == 0 {
		return true
	} else {
		return false
	}
}

func handleDirectory(directoryPath string) {
	// Read all files in directory
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Sort better
	// // Save each file path in directory
	// var filepaths []string
	// for _, f := range files {
	// 	if !f.Mode().IsDir() {
	// 		filepaths = append(filepaths, path.Join(directoryPath, f.Name()))
	// 	}
	// }
	// sort.Strings(filepaths)

	// // Go through and deal with each file
	// for _, path := range filepaths {
	// 	handleIndividualFile(path)
	// }

	// Load data of all png we find
	for _, file := range files {
		if !file.Mode().IsDir() {
			filePath := path.Join(directoryPath, file.Name())
			handleIndividualFile(filePath)
		}
	}
}

func handleIndividualFile(filePath string) {
	// Check file type
	if isImageFile(filePath, "png") {

		// Try to open file
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Get image object data
		imgData, _, err := image.Decode(file)
		if err != nil {
			log.Fatal(err)
		}

		// Load additional image data and store in array
		imgEntry, err := extractPNGData(filePath, imgData)
		if err != nil {
			log.Fatal(err)
		}
		images = append(images, &imgEntry)

		fileCount++
		log.Printf("Loaded image [%d]: %s: %dx%d", fileCount, imgEntry.path, imgEntry.width, imgEntry.height)

	} else if isImageFile(filePath, "gif") {
		extractGIFData(filePath)
	} else {
		log.Fatal("File type not supported.")
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
