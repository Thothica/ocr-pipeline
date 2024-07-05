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
	PDF_DIR             = "openalex-pdfs/"
	MAX_ROUTINES        = 3000
	OPENALEX_NEW        = "extracted.jsonl.gz"
)

var (
	bad_files       atomic.Int64
	wg              sync.WaitGroup
	limitChannel    = make(chan struct{}, MAX_ROUTINES)
	writeObjChannel = make(chan *openalex.WorkObject, MAX_ROUTINES)
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
		limitChannel <- struct{}{}
		wg.Add(1)
		go func(oj openalex.WorkObject) {
			err := oj.ExtractTextFromPdf(PDF_DIR, writeObjChannel)
			if err != nil {
				bad_files.Add(1)
				color.Red(fmt.Sprintf("Failed: %s", err))
			}
			<-limitChannel
			wg.Done()
		}(obj)
	}
	close(limitChannel)

	wg.Add(1)
	go func() {
		file, err := os.Create(OPENALEX_NEW)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		writer := gzip.NewWriter(file)
		for msg := range writeObjChannel {
			objBytes, err := json.Marshal((*msg))
			if err != nil {
				bad_files.Add(1)
				color.Red(fmt.Sprintf("Failed to marshal: %s", err))
			}
			objBytes = append(objBytes, '\n')
			writer.Write(objBytes)
		}
		writer.Close()
		writer.Flush()
		wg.Done()
	}()

	close(writeObjChannel)
	wg.Wait()
	color.White("All done !!")
	color.Red("FILES LOST: %v", bad_files.Load())
	color.Green("FILES EXTRACTED: %v", total-int(bad_files.Load()))
}
