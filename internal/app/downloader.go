package app

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

func (d SimpleImageDownloader) Download(ctx context.Context, url string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	return d.client.Do(req)
}
