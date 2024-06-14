package main

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetLanguage("ara")
	client.SetImage(".png")
	text, err := client.Text()
	if err != nil {
		panic(err)
	}
	fmt.Println(text)
}

