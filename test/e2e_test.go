package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/bardex/minipic/internal"
	"github.com/bardex/minipic/pkg/respcache"

	"github.com/stretchr/testify/require"
)

func newImageServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vs := range r.Header {
			for _, v := range vs {
				w.Header().Add("X-From-"+k, v)
			}
		}
		w.Header().Add("X-Name", "test-server")
		w.Header().Add("X-Custom2", "test-server")

		switch r.RequestURI {
		case "/sample.jpg":
			http.ServeFile(w, r, "sample.jpg")
		case "/html":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("HTML"))
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
	h := internal.NewHandler(
		internal.NewImageDownloader(),
		internal.ResizerByImaging{},
	)
	h = respcache.NewCacheMiddleware(respcache.NewLruCache("tmp", 2), h)
	return httptest.NewServer(h)
}

func TestImageServerWorked(t *testing.T) {
	is := newImageServer()
	defer is.Close()

	res, err := http.Get(is.URL + "/sample.jpg")
	defer res.Body.Close()

	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)

	f, err := os.Open("sample.jpg")
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
