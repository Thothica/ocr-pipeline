package openalex

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/ledongthuc/pdf"
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{
		Transport: tr,
	}
)

func (obj *WorkObject) SaveArticle(downloadDir string) error {
	resp, err := client.Get(obj.OpenAccess.OAURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if len(resp.Header["Content-Type"]) < 1 {
		return errors.New("Invalid response")
	}

	if resp.Header["Content-Type"][0] != "application/pdf" {
		return err
	}

	title := fmt.Sprintf("%s.%s", strings.Split(obj.Id, "/")[3], "pdf")
	file, err := os.Create(filepath.Join(downloadDir, title))
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

func (obj *WorkObject) ExtractTextFromPdf(pdfDir string, savePdfChannel chan<- *WorkObject) error {
	title := fmt.Sprintf("%s.%s", strings.Split(obj.Id, "/")[3], "pdf")
	_, err := os.Stat(filepath.Join(pdfDir, title))
	if err != nil {
		return err
	}
	f, r, err := pdf.Open(filepath.Join(pdfDir, title))
	defer f.Close()
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return err
	}
	buf.ReadFrom(b)
	obj.Text = buf.String()
	savePdfChannel <- obj
	return nil
}
