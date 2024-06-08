package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

type ImageConvert struct {
	title   string
	pageNum int
	Image   bytes.Buffer
}

const (
	BOOKS_DIR    = "books/"
	MAX_ROUTINES = 100
)

var (
	conversionChannel = make(chan *ImageConvert, MAX_ROUTINES)
)

func main() {
	err := filepath.WalkDir(BOOKS_DIR, func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(d.Name(), ".pdf") {
			return nil
		}
		err = ConvertPdf(path)
		if err != nil {

		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}

func ConvertPdf(path string) error {
	doc, err := fitz.New(path)
	if err != nil {
		return err
	}

	for n := 0; n < doc.NumPage(); n++ {
		var CurrImage bytes.Buffer

		img, err := doc.Image(n)
		if err != nil {
			return err
		}

		err = jpeg.Encode(&CurrImage, img, &jpeg.Options{Quality: jpeg.DefaultQuality})

		i := &ImageConvert{
			title:   path,
			pageNum: n,
			Image:   CurrImage,
		}
		conversionChannel <- i
	}

	return nil
}
