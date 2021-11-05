package internal

import (
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

func (d SimpleImageDownloader) Download(URL string, headers http.Header) (*http.Response, error) {
	//TODO: separate errors by types
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = headers
	//TODO: validate - is the image
	return d.client.Do(req)
}
