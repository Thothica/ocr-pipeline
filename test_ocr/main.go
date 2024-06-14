package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/otiai10/gosseract/v2"
)

func main() {
	var bookText strings.Builder
	client := gosseract.NewClient()
	defer client.Close()
	client.Languages = []string{"ara"}
	client.SetImage("image-0100.jpg")
	text, err := client.Text()
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&bookText, "%s\n\n", text)

	file, err := os.Create("test.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(bookText.String())
	if err := file.Sync(); err != nil {
		panic(err)
	}
	color.Green(fmt.Sprintf("Completed: "))
}
