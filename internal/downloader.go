package internal

import (
	"context"
	"net/http"
	"time"
)

type SimpleImageDownloader struct {
	client *http.Client
}

func NewImageDownloader() SimpleImageDownloader {
	return SimpleImageDownloader{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 30 * time.Second,
			},
		},
	}
}

func (d SimpleImageDownloader) Download(url string, headers http.Header) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	return d.client.Do(req)
}
