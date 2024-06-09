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
	BOOKS_DIR           = "books/"
	OUTPUT_DIR          = "images/"
	MAX_PDF_CONVERSIONS = 100
)

var (
	wg       sync.WaitGroup
	badFiles atomic.Int64
	ch       = make(chan struct{}, MAX_PDF_CONVERSIONS)
)

func main() {
	err := filepath.WalkDir(BOOKS_DIR, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(d.Name(), ".pdf") {
			return nil
		}

		wg.Add(1)
		ch <- struct{}{}
		go func() {
			err = ConvertPdf(path)
			if err != nil {
				badFiles.Add(1)
				color.Red(fmt.Sprintf("Failed: %s", d.Name()))
			}
			<-ch
			wg.Done()
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
	curDir := filepath.Join(OUTPUT_DIR, title)
	lenPages := doc.NumPage()
	err = os.MkdirAll(curDir, 0750)
	if err != nil {
		panic(err)
	}

	for n := 0; n < lenPages; n++ {
		img, err := doc.Image(n)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(curDir, fmt.Sprintf("image-%03d.jpg", n)))
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
