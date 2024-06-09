package main

import (
	"fmt"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/gen2brain/go-fitz"
)

const (
	BOOKS_DIR  = "books/"
	OUTPUT_DIR = "images/"
	// MAX_PDF_CONVERSIONS = 100
)

var (
	wg       sync.WaitGroup
	badFiles atomic.Int64
)

func main() {
	err := filepath.WalkDir(BOOKS_DIR, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(d.Name(), ".pdf") {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = ConvertPdf(path)
			if err != nil {
				badFiles.Add(1)
				color.Red(fmt.Sprintf("Failed: %s", d.Name()))
			}
		}()

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	wg.Wait()
	color.White("Completed: \n\t Books Failed: %v", badFiles.Load())
}
func ConvertPdf(path string) error {
	doc, err := fitz.New(path)
	if err != nil {
		return err
	}
	defer doc.Close()

	title := strings.Split(strings.Split(path, ".")[0], "/")[1]
	lenPages := doc.NumPage()
	tmpDir, err := os.MkdirTemp(OUTPUT_DIR, title)
	if err != nil {
		panic(err)
	}

	for n := 0; n < lenPages; n++ {
		img, err := doc.Image(n)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("image-%03d.jpg", n)))
		if err != nil {
			panic(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			return err
		}

		f.Close()
	}

	color.Green(fmt.Sprintf("Completed: %s", title))

	return nil
}
