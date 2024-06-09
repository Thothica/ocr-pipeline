package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
)

type ImageConvert struct {
	title  string
	images []bytes.Buffer
}

const (
	BOOKS_DIR           = "books/"
	MAX_PDF_CONVERSIONS = 100
	MAX_READERS         = 50
	OUTPUT_DIR          = "texts/"
)

var (
	conversionChannel = make(chan *ImageConvert, MAX_PDF_CONVERSIONS)
	wg                sync.WaitGroup
)

func main() {
	for range MAX_READERS {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ExtractedImages := <-conversionChannel

			var b strings.Builder
			client := gosseract.NewClient()
			defer client.Close()

			for _, image := range ExtractedImages.images {
				client.SetImageFromBytes(image.Bytes())
				text, err := client.Text()
				if err != nil {
					color.Red("Failed ocr: %v \n book: %s", err, ExtractedImages.title)
					return
				}
				fmt.Fprintf(&b, "%s\n\n", text)
			}

			pathToSave := filepath.Join(OUTPUT_DIR, fmt.Sprintf("%s.txt", ExtractedImages.title))
			file, err := os.Create(pathToSave)
			if err != nil {
				color.Red("Failed to create txt file: %v \n book: %s", err, ExtractedImages.title)
				return
			}
			defer file.Close()

			_, err = fmt.Fprintf(file, b.String())
			if err != nil {
				color.Red("Failed to write txt file: %v \n book: %s", err, ExtractedImages.title)
			}
		}()
	}

	err := filepath.WalkDir(BOOKS_DIR, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(d.Name(), ".pdf") {
			return nil
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = ConvertPdf(path)
			if err != nil {
				color.Red(fmt.Sprintf("Failed: %s", d.Name()))
			}
		}()

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

	wg.Wait()
}

func ConvertPdf(path string) error {
	doc, err := fitz.New(path)
	if err != nil {
		return err
	}

	lenPages := doc.NumPage()
	images := make([]bytes.Buffer, lenPages)

	for n := 0; n < lenPages; n++ {
		img, err := doc.Image(n)
		if err != nil {
			return err
		}

		err = jpeg.Encode(&images[n], img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			return err
		}
	}

	i := &ImageConvert{
		title:  path,
		images: images,
	}
	conversionChannel <- i

	return nil
}
