package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	image := flag.String("image", "", "image url")
	flag.Parse()

	_, err := url.ParseRequestURI(*image)
	if err != nil {
		panic(err)
	}

	if err := download(*image); err != nil {
		// log something
		panic(err)
	}

}

func download(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New(http.StatusText(res.StatusCode))
	}

	filepath := path.Base(res.Request.URL.String())
	nf, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer nf.Close()

	pw := &PrintWriter{
		Total: uint64(res.ContentLength),
	}
	_, err = io.Copy(nf, io.TeeReader(res.Body, pw))

	return err
}

type PrintWriter struct {
	Total      uint64
	Downloaded uint64
}

func (pw *PrintWriter) Write(p []byte) (n int, err error) {
	pw.Downloaded = uint64(len(p)) + pw.Downloaded
	pw.Print()

	return len(p), nil
}

const progressBarLength = 50

var (
	Green  = Color("\033[32m%s\033[0m")
	Yellow = Color("\033[33m%s\033[0m")
	Red    = Color("\033[31m%s\033[0m")
)

func Color(colorStr string) func(...interface{}) string {
	return func(args ...interface{}) string {
		return fmt.Sprintf(colorStr, fmt.Sprint(args...))
	}
}

func (pw *PrintWriter) Print() {
	progressBarCompleted := int(progressBarLength * pw.Downloaded / pw.Total)
	progressBarLeft := strings.Repeat("~", progressBarLength-progressBarCompleted)
	progress := strings.Repeat("█", progressBarCompleted)

	fmt.Printf("\r %s [%s%s] %.1f", Red("Downloading"), Green(progress), Yellow(progressBarLeft), float64(pw.Downloaded)/float64(pw.Total)*100)
	time.Sleep(200 * time.Millisecond)
}
