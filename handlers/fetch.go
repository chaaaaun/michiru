package handlers

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"michiru/config"
)

func FetchDump(ctx context.Context, cfg config.Config) ([]byte, error) {
	url := cfg.TitleDumpURL

	logger.Println("Fetching dump from ", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("User-Agent", "go-http")

	client := &http.Client{Timeout: cfg.FetchTimeout}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching title dump: %w", err)
	}
	defer func(res *http.Response) {
		cerr := res.Body.Close()
		if cerr != nil {
			logger.Println("warn: ", cerr)
		}
	}(res)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching title dump: %w", err)
	}

	zr, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("decompressing dump: %w", err)
	}
	defer func(zr *gzip.Reader) {
		cerr := zr.Close()
		if cerr != nil {
			logger.Println("warn: ", cerr)
		}
	}(zr)

	b, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("reading dump file: %w", err)
	}

	return b, nil
}

func FetchDumpMock(ctx context.Context, cfg config.Config) ([]byte, error) {
	logger.Println("Reading dump from", "anime-titles.xml.gz")
	file, err := os.Open("anime-titles.xml.gz")
	if err != nil {
		return nil, fmt.Errorf("opening dump file: %w", err)
	}
	defer func(file *os.File) {
		cerr := file.Close()
		if cerr != nil {
			logger.Println("warn: ", err)
		}
	}(file)

	zr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("decompressing dump: %w", err)
	}
	defer func(zr *gzip.Reader) {
		cerr := zr.Close()
		if cerr != nil {
			logger.Println("warn: ", err)
		}
	}(zr)

	b, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("reading dump file: %w", err)
	}

	return b, nil
}
