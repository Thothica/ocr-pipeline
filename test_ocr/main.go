package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const (
	IMAGES_DIR         = "images/"
	OUTPUT_DIR         = "texts/"
	MAX_PARALLEL_BOOKS = 100
)

func main() {
	books, err := os.ReadDir(IMAGES_DIR)
	if err != nil {
		panic(err)
	}

	for _, book := range books {
		if book.IsDir() {
			err := extractBook(filepath.Join(IMAGES_DIR, book.Name()))
			if err != nil {
				color.Red(fmt.Sprintf("Failed: %s", book.Name()))
			}
		}
	}
}

func extractBook(bookDir string) error {
	_, err := os.ReadDir(bookDir)
	if err != nil {
		return err
	}

	title := strings.Split(bookDir, "/")[1]
	fmt.Println(title)

	return nil
}
