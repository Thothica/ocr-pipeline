package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"

	"gopkg.in/gographics/imagick.v2/imagick"
)

const (
	BOOKS_DIR           = "books/"
	OUTPUT_DIR          = "images/"
	MAX_PDF_CONVERSIONS = 1000
)

var (
	wg       sync.WaitGroup
	badFiles atomic.Int64
	ch       = make(chan struct{}, MAX_PDF_CONVERSIONS)
)

func main() {
	imagick.Initialize()
	defer imagick.Terminate()

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
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	mw.ReadImage(path)
	if err := mw.SetFormat("jpg"); err != nil {
		log.Fatal(err)
	}

	title := strings.Split(strings.Split(path, ".")[0], "/")[1]
	curDir := filepath.Join(OUTPUT_DIR, title)
	err := os.MkdirAll(curDir, 0750)
	if err != nil {
		panic(err)
	}

	page := 0

	for {
		if ok := mw.SetIteratorIndex(page); !ok {
			break
		}

		if err := mw.WriteImage(filepath.Join(curDir, fmt.Sprintf("image-%04d.jpg", page))); err != nil {
			return err
		}

		page += 1
	}

	color.Green(fmt.Sprintf("Completed: %s", title))
	return nil
}
