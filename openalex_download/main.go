package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/Thothica/ocr-pipeline/internal/openalex"
	"github.com/fatih/color"
)

const (
	OPENALEX_BEST_WORKS = "best_works.jsonl.gz"
	DOWNLOAD_DIR        = "openalex-pdfs/"
	MAX_ROUTINES        = 3000
)

var (
	bad_files      atomic.Int64
	wg             sync.WaitGroup
	routineChannel = make(chan struct{}, MAX_ROUTINES)
)

func main() {
	file, err := os.Open(OPENALEX_BEST_WORKS)
	if err != nil {
		panic(err)
	}

	rawContents, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(rawContents)

	total := 0
	for scanner.Scan() {
		var obj openalex.WorkObject
		text := scanner.Text()
		err := json.Unmarshal([]byte(text), &obj)
		if err != nil {
			bad_files.Add(1)
			continue
		}
		total += 1
		routineChannel <- struct{}{}
		wg.Add(1)
		go func(oj openalex.WorkObject) {
			err := oj.SaveArticle(DOWNLOAD_DIR)
			if err != nil {
				bad_files.Add(1)
				color.Red(fmt.Sprintf("Failed: %s", err))
			}
			<-routineChannel
			wg.Done()
		}(obj)
	}

	wg.Wait()
	color.White("All done !!")
	color.Red("FILES LOST: %v", bad_files.Load())
	color.Green("FILES DOWNLOADED: %v", total-int(bad_files.Load()))
}
