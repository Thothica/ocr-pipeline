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
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

const (
	CSV_FILE      = "url.csv"
	TARGET_DIR    = "books/"
	INITIAL_INDEX = 469
	MAX_ROUTINES  = 1500
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

	for idx, val := range csvData {
		routineChannel <- struct{}{}
		wg.Add(1)
		go func(url, title string) {
			err := download_pdf(url, title)
			if err != nil {
				bad_files.Add(1)
				color.Red(fmt.Sprintf("Failed: %s", title))
			}
			<-routineChannel
			wg.Done()
		}(val[5], fmt.Sprintf("%v", idx+INITIAL_INDEX))
	}

	wg.Wait()
	color.White("All done !!")
	color.Red("FILES LOST: %v", bad_files.Load())
	color.Green("FILES DOWNLOADED: %v", len(csvData)-int(bad_files.Load()))
}

func download_pdf(url, title string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.Header["Content-Type"][0] != "application/pdf" {
		return err
	}

	title = strings.ReplaceAll(title, "/", "-")
	file_path := filepath.Join(TARGET_DIR, fmt.Sprintf("%s%s", title, ".pdf"))

	file, err := os.Create(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	color.Green(fmt.Sprintf("Downloaded: %s", title))
	return nil
}
