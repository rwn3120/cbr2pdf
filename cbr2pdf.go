package main

import (
	"errors"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gen2brain/go-unarr"
	"github.com/nfnt/resize"
	"github.com/signintech/gopdf"
)

type resolution struct {
	w uint
	h uint
}

var defaultWidth uint = 1072
var defaultHeight uint = 1448

var pageResolution = resolution{
	getEnv("WIDTH", defaultWidth),
	getEnv("HEIGHT", defaultHeight)}

func getEnv(key string, defaultValue uint) uint {
	value, found := os.LookupEnv(key)
	if !found {
		return defaultValue
	}
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		fmt.Printf("%s has invalid value: %s\n", key, value)
		os.Exit(253)
	}
	return uint(v)
}

func imgToPdf(src string, pdf *gopdf.GoPdf) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}
	file.Close()
	// resize to width 1000 using Lanczos resampling and preserve aspect ratio
	if img.Bounds().Dx() < img.Bounds().Dy() {
		img = resize.Resize(0, pageResolution.h, img, resize.Lanczos3)
	} else {
		img = resize.Resize(pageResolution.h, 0, img, resize.Lanczos3)
	}
	// add page
	pageSize := &gopdf.Rect{
		W: float64(img.Bounds().Dx()) / 1.78,
		H: float64(img.Bounds().Dy()) / 1.78}
	pdf.AddPageWithOption(gopdf.PageOption{PageSize: pageSize})
	// store file
	src = src + ".resized"
	out, err := os.Create(src)
	if err != nil {
		return err
	}
	defer out.Close()
	// write resized image to file
	if err := jpeg.Encode(out, img, nil); err != nil {
		return err
	}
	if err := pdf.Image(src, 0, 0, nil); err != nil {
		return err
	}
	return nil
}

func unrar(src, dst string) error {
	a, err := unarr.NewArchive(src)
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Extract(dst)
}

func findImages(location string) ([]string, error) {
	images := []string{}
	err := filepath.Walk(location,
		func(file string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.EqualFold(".jpg", path.Ext(file)) || strings.EqualFold(".jpeg", path.Ext(file)) {
				images = append(images, file)
			}
			return nil
		})
	sort.Slice(images, func(i, j int) bool {
		return images[i] < images[j]
	})
	return images, err
}

// Convert converts cbr to pdf
func Convert(src, dst string) error {
	if src == dst {
		return errors.New("Source file must not be equal to destination file")
	}
	// extract files
	tmpDir, err := ioutil.TempDir(os.TempDir(), filepath.Base(src))
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir) // clean up
	if err := unrar(src, tmpDir); err != nil {
		return err
	}
	// find images
	images, err := findImages(tmpDir)
	if err != nil {
		return err
	}
	if len(images) < 1 {
		return fmt.Errorf("%s is empty", src)
	}
	// create pdf
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{})
	// add images to pdf
	for index, image := range images {
		fmt.Printf("Processing image %s (%d/%d)...\r", image, index+1, len(images))
		if err := imgToPdf(image, pdf); err != nil {
			fmt.Printf("\nFailed to add image %s (%d/%d)!\n", image, index+1, len(images))
			return err
		}
	}
	// create pdf
	fmt.Printf("\nPlease wait - generating %s\n", dst)
	return pdf.WritePdf(dst)
}

func parseArgs(args []string) (string, string, error) {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "-?" {
			filename := filepath.Base(args[0])
			fmt.Printf(
				"%s - an utility for converting CBR/CBZ to PDF.\n\n"+
					"Usage: %s <source file> [destination file]\n\n"+
					"Environment:    WIDTH  ... X resolution of your reader (default %d)\n"+
					"                HEIGHT ... Y resolution of your reader (default %d)\n\n\n"+
					"Examples: \n"+
					"          ./%s my-favorite-comicbook.cbr output.pdf                             # Pocketbook Touch HD 3\n"+
					"          WIDTH=758 HEIGHT=1024 ./%s my-favorite-comicbook.cbr output.pdf       # Pocketbook Touch Lux 4\n\n",
				filename, filename, defaultWidth, defaultHeight, filename, filename)
			os.Exit(1)
		}
	}

	if len(os.Args) < 2 {
		return "", "", errors.New("Missing mandatory argument: source file\nRun with --help to display usage")
	}
	src := args[1]
	if len(args) > 2 {
		return src, args[2], nil
	}
	ext := path.Ext(src)
	if ext == "" {
		return src, src + ".pdf", nil
	}
	return src, src[0:len(src)-len(ext)] + ".pdf", nil
}

func main() {
	src, dst, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println("An error has occurred:", err)
		os.Exit(255)
	}
	if err := Convert(src, dst); err != nil {
		fmt.Println("Conversion failed:", err)
		os.Exit(254)
	}
}
