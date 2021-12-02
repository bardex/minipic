package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bardex/minipic/internal/app"
	"github.com/stretchr/testify/require"
)

func TestResponseCache(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		body    []byte
	}{
		{
			name:    "standard",
			headers: http.Header{"Content-Type": {"image/png"}, "Transfer-Encoding": {"chunked"}, "Date": {"Fri, 26 Nov 2021 11:06:48 GMT"}},
			body:    []byte{32, 33, 34, 35},
		},
		{
			name:    "no headers",
			headers: nil,
			body:    []byte{32},
		},
		{
			name:    "large headers and body",
			headers: http.Header{"Content-Type": {"image/png"}, "X-Test": {strings.Repeat("s", 1000)}},
			body:    bytes.Repeat([]byte{32}, 1000000),
		},
	}

	cache := app.NewLruCache("/tmp", 5)
	defer cache.Clear()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rc := cache.CreateItem(tt.name)

			err := rc.Save(tt.headers, tt.body)
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			err = rc.WriteTo(rw)
			require.NoError(t, err)

			res := rw.Result()
			defer res.Body.Close()

			if tt.headers != nil {
				for k, v := range tt.headers {
					require.Equal(t, v[0], res.Header.Get(k))
				}
			}
			require.Equal(t, tt.body, rw.Body.Bytes())
			err = rc.Remove()
			require.NoError(t, err)
		})
	}
}
