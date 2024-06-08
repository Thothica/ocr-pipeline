package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

const (
	CSV_FILE      = "url.csv"
	TARGET_DIR    = "books/"
	INITIAL_INDEX = 469
	MAX_ROUTINES  = 400
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{
		Transport: tr,
	}
	bad_files      atomic.Int64
	good_files     atomic.Int64
	wg             sync.WaitGroup
	routineChannel = make(chan struct{}, MAX_ROUTINES)
)

func main() {
	file, err := os.Open(CSV_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	for range INITIAL_INDEX + 1 {
		reader.Read()
	}

	csvData, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, val := range csvData {
		routineChannel <- struct{}{}
		wg.Add(1)
		go func(url, title string) {
			download_pdf(url, title)
			<-routineChannel
			wg.Done()
		}(val[5], val[1])
	}

	wg.Wait()
	color.White("All done !!")
	color.Red("FILES LOST: %v", bad_files.Load())
	color.Green("FILES DOWNLOADED: %v", len(csvData)-int(bad_files.Load()))
}

func download_pdf(url, title string) {
	resp, err := client.Get(url)
	if err != nil {
		bad_files.Add(1)
		color.Red(fmt.Sprintf("Failed: %s", title))
		return
	}
	defer resp.Body.Close()

	if resp.Header["Content-Type"][0] != "application/pdf" {
		bad_files.Add(1)
		color.Red(fmt.Sprintf("Failed: %s", title))
		return
	}

	file_path := filepath.Join(TARGET_DIR, fmt.Sprintf("%s%s", title, ".pdf"))
	file, err := os.Create(file_path)
	if err != nil {
		bad_files.Add(1)
		color.Red(fmt.Sprintf("Failed: %s", title))
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		bad_files.Add(1)
		color.Red(fmt.Sprintf("Failed: %s", title))
		return
	}

	color.Green(fmt.Sprintf("Downloaded: %s", title))
}
