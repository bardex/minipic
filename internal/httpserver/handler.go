package httpserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// ResizeModeFit fit image into the specified sizes.
	ResizeModeFit = "fit"

	// ResizeModeFill fill given dimensions with image.
	ResizeModeFill = "fill"
)

type Downloader interface {
	Download(ctx context.Context, URL string, headers http.Header) (*http.Response, error)
}

type ImageResizer interface {
	Resize(src io.Reader, dst io.Writer, opts ResizeOptions) error
}

type ResizeOptions struct {
	Mode   string
	Width  int
	Height int
}

type Handler struct {
	downloader Downloader
	resizer    ImageResizer
}

func NewHandler(d Downloader, r ImageResizer) http.Handler {
	return Handler{
		downloader: d,
		resizer:    r,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	src, opts, err := h.parseRequestURI(r.URL.RequestURI())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.downloader.Download(ctx, src, r.Header.Clone())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	res.Header.Del("Content-Length")

	for k, v := range res.Header {
		w.Header()[k] = v
	}

	if res.StatusCode != 200 {
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
		return
	}

	var img bytes.Buffer
	if err = h.resizer.Resize(res.Body, &img, opts); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(img.Len()))
	io.Copy(w, &img)
}

func (h Handler) parseRequestURI(uri string) (src string, opts ResizeOptions, err error) {
	uri = strings.Trim(uri, "/")
	params := strings.SplitN(uri, "/", 4)
	if len(params) != 4 {
		err = errors.New("request URL should look like /<mode>/<width>/<height>/<image_url>")
		return
	}
	mode := params[0]
	if mode != ResizeModeFill && mode != ResizeModeFit {
		err = fmt.Errorf("resize mode must be `%s` or `%s`", ResizeModeFill, ResizeModeFit)
		return
	}
	width, err := strconv.Atoi(params[1])
	if err != nil {
		err = errors.New("image width must be integer")
		return
	}
	height, err := strconv.Atoi(params[2])
	if err != nil {
		err = errors.New("image height must be integer")
		return
	}
	if width < 10 || height < 10 {
		err = errors.New("width and height must be more than 10px")
		return
	}

	imgSrc, err := url.ParseRequestURI(params[3])
	if err != nil || (imgSrc.Scheme != "http" && imgSrc.Scheme != "https") || imgSrc.Host == "" {
		err = errors.New("image URL must be absolute")
		return
	}

	src = params[3]
	opts.Mode = mode
	opts.Width = width
	opts.Height = height

	return
}
