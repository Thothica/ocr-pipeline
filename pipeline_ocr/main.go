package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/otiai10/gosseract/v2"
)

const (
	IMAGES_DIR         = "images/"
	OUTPUT_DIR         = "texts/"
	MAX_PARALLEL_BOOKS = 100
)

var (
	wg       sync.WaitGroup
	badFiles atomic.Int64
	ch       = make(chan struct{}, MAX_PARALLEL_BOOKS)
)

func main() {
	books, err := os.ReadDir(IMAGES_DIR)
	if err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		if book.IsDir() {
			err := extractBook(filepath.Join(IMAGES_DIR, book.Name()))
			if err != nil {
				color.Red(fmt.Sprintf("Failed: %s", book.Name()))
				badFiles.Add(1)
			}
		}
	}
}

func extractBook(bookDir string) error {
	images, err := os.ReadDir(bookDir)
	if err != nil {
		return err
	}

	var bookText strings.Builder
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("ara")
	title := strings.Split(bookDir, "/")[1]

	for _, image := range images {
		client.SetImage(filepath.Join(bookDir, image.Name()))
		text, err := client.Text()
		if err != nil {
			return err
		}
		fmt.Fprintf(&bookText, "%s\n\n", text)
	}

	file, err := os.Create(filepath.Join(OUTPUT_DIR, title))
	if err != nil {
		return err
	}
    defer file.Close()

	w := bufio.NewWriter(file)
	w.WriteString(bookText.String())
	if err := w.Flush(); err != nil {
		return err
	}

	color.Green(fmt.Sprintf("Completed: %s", title))
	return nil
}
