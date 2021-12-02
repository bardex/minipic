package test

import (
	"bytes"
	"context"
	"image"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bardex/minipic/internal/app"
	"github.com/bardex/minipic/internal/httpserver"
	"github.com/bardex/minipic/internal/httpserver/middleware"
	"github.com/stretchr/testify/require"
)

func newImageServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vs := range r.Header {
			for _, v := range vs {
				// return request headers for check it
				w.Header().Add("X-From-"+k, v)
			}
		}
		w.Header().Add("X-Name", "test-server")
		w.Header().Add("X-Server", "cloud")

		switch r.RequestURI {
		case "/sample.jpeg":
			http.ServeFile(w, r, "sample.jpeg")
		case "/sample.png":
			http.ServeFile(w, r, "sample.png")
		case "/sample.webp":
			http.ServeFile(w, r, "sample.webp")
		case "/500":
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("x-error", "500")
			w.Write([]byte("500 Server Error"))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("x-error", "404")
			w.Write([]byte("404 Not Found"))
		}
	}))
}

func newMinipicServer() *httptest.Server {
	h := httpserver.NewHandler(
		app.NewImageDownloader(),
		app.Resizer{},
	)
	h = middleware.NewCache(app.NewLruCache("tmp", 2), h)
	return httptest.NewServer(h)
}

func TestLocalImageServer(t *testing.T) {
	is := newImageServer()
	defer is.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, is.URL+"/sample.jpeg", nil)
	require.NoError(t, err)
	var client http.Client
	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, 200, res.StatusCode)

	f, err := os.Open("sample.jpeg")
	require.NoError(t, err)
	finfo, err := f.Stat()
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(finfo.Size(), 10), res.Header.Get("Content-Length"))
	require.Equal(t, "image/jpeg", res.Header.Get("Content-Type"))

	imgSer, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	imgLoc, err := io.ReadAll(f)
	require.NoError(t, err)
	require.True(t, bytes.Equal(imgSer, imgLoc))
}

func TestMinipicServer(t *testing.T) {
	is := newImageServer()
	defer is.Close()
	mp := newMinipicServer()
	defer mp.Close()

	tests := []struct {
		url    string
		status int
		w      int
		h      int
	}{
		{url: mp.URL + "/fill/500/500/" + is.URL + "/sample.png", status: 200, w: 500, h: 500},
		{url: mp.URL + "/fit/800/600/" + is.URL + "/sample.png", status: 200, w: 800, h: 600},
		{url: mp.URL + "/fill/500/500/" + is.URL + "/sample.jpeg", status: 200, w: 500, h: 500},
		{url: mp.URL + "/fit/800/800/" + is.URL + "/sample.jpeg", status: 200, w: 800, h: 800},
		{url: mp.URL + "/fit/800/800/" + is.URL + "/sample.webp", status: 502, w: 800, h: 800},
		{url: mp.URL + "/fit/800/800/" + is.URL + "/404", status: 404, w: 800, h: 800},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.url, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tt.url, nil)
			require.NoError(t, err)
			req.Header.Set("User-Agent", "Firefox")
			req.Header.Set("Accept-Encoding", "gzip,deflate")

			var client http.Client
			result, err := client.Do(req)

			require.NoError(t, err)
			defer result.Body.Close()

			require.Equal(t, tt.status, result.StatusCode)
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(len(body)), result.Header.Get("Content-Length"))

			// tests proxy-headers
			require.Equal(t, "test-server", result.Header.Get("X-Name"))
			require.Equal(t, req.Header.Get("User-Agent"), result.Header.Get("X-From-User-Agent"))
			require.Equal(t, req.Header.Get("Accept-Encoding"), result.Header.Get("X-From-Accept-Encoding"))

			if result.StatusCode == 200 {
				require.Contains(t, result.Header.Get("Content-Type"), "image/")
				img, _, err := image.Decode(bytes.NewReader(body))
				require.NoError(t, err)
				w := img.Bounds().Max.X
				h := img.Bounds().Max.Y
				require.LessOrEqual(t, w, tt.w)
				require.LessOrEqual(t, h, tt.h)
			}
		})
	}
}
