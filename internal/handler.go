package internal

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// ResizeModeFit fit image into the specified sizes.
	ResizeModeFit = "fit"
	// ResizeModeFill fill given dimensions with image.
	ResizeModeFill = "fill"
)

type ImageDownloader interface {
	Download(URL string, headers http.Header) (*http.Response, error)
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
	downloader ImageDownloader
	resizer    ImageResizer
}

func NewHandler(d ImageDownloader, r ImageResizer) http.Handler {
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

	res, err := h.downloader.Download(src, r.Header.Clone())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)
	for k, v := range res.Header {
		// because resize...
		if strings.ToLower(k) == "content-length" {
			continue
		}
		w.Header()[k] = v
	}

	// check error
	if res.StatusCode != 200 {
		if _, err := io.Copy(w, res.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = h.resizer.Resize(res.Body, w, opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
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
		err = errors.New("invalid image url")
		return
	}

	src = params[3]
	opts.Mode = mode
	opts.Width = width
	opts.Height = height

	return
}
