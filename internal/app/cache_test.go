package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponseCache(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		body    []byte
	}{
		{
			name: "standard",
			headers: http.Header{
				"Content-Type":      {"image/png"},
				"Transfer-Encoding": {"chunked"},
				"Date":              {"Fri, 26 Nov 2021 11:06:48 GMT"},
			},
			body: []byte{32, 33, 34, 35},
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

	cache := NewLruCache("/tmp", 1)
	defer cache.Clear()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Save(tt.name, tt.headers, tt.body)
			require.NoError(t, err)
			rw := httptest.NewRecorder()
			hit, err := cache.GetAndWriteTo(tt.name, rw)
			require.NoError(t, err)
			require.True(t, hit)

			res := rw.Result()
			defer res.Body.Close()

			if tt.headers != nil {
				for k, v := range tt.headers {
					require.Equal(t, v[0], res.Header.Get(k))
				}
			}
			require.Equal(t, tt.body, rw.Body.Bytes())
		})
	}
}

func TestLRUCache_Concurrency(t *testing.T) {
	t.Parallel()
	cCap := 3
	cache := NewLruCache("/tmp", cCap)
	t.Cleanup(func() {
		cache.Clear()
	})

	items := []struct {
		key      string
		headers  http.Header
		body     []byte
		filepath string
	}{
		{key: "1", headers: http.Header{"Content-Type": {"image/png"}}, body: []byte{10, 10, 10, 10}},
		{key: "2", headers: http.Header{"Content-Type": {"image/jpeg"}}, body: []byte{20, 20, 20}},
		{key: "3", headers: http.Header{"Content-Type": {"image/bmp"}}, body: []byte{30, 30}},
		{key: "4", headers: http.Header{"Content-Type": {"image/gif"}}, body: []byte{40, 40}},
		{key: "5", headers: http.Header{"Content-Type": {"image/webp"}}, body: []byte{50, 50, 50}},
		{key: "6", headers: http.Header{"Content-Type": {"image/tiff"}}, body: []byte{50, 50, 50}},
		{key: "7", headers: http.Header{"Content-Type": {"image/jpg"}}, body: []byte{50, 50, 50}},
		{key: "8", headers: http.Header{"Content-Type": {"image/svg"}}, body: []byte{50, 50, 50}},
		{key: "9", headers: http.Header{"Content-Type": {"image/psd"}}, body: []byte{50, 50, 50}},
	}

	for _, item := range items {
		item := item
		for i := 0; i < 5; i++ {
			t.Run(fmt.Sprintf("%s_%d", item.key, i), func(t *testing.T) {
				t.Parallel()

				err := cache.Save(item.key, item.headers, item.body)
				require.NoError(t, err)

				rw := httptest.NewRecorder()
				_, err = cache.GetAndWriteTo(item.key, rw)
				require.NoError(t, err)
			})
		}
	}
}

func TestLRUCache(t *testing.T) {
	cCap := 3
	cache := NewLruCache("/tmp", cCap)
	defer cache.Clear()

	items := []struct {
		key     string
		headers http.Header
		body    []byte
	}{
		{key: "1", headers: http.Header{"Content-Type": {"image/png"}}, body: []byte{10, 10, 10, 10}},
		{key: "2", headers: http.Header{"Content-Type": {"image/jpeg"}}, body: []byte{20, 20, 20}},
		{key: "3", headers: http.Header{"Content-Type": {"image/bmp"}}, body: []byte{30, 30}},
		{key: "4", headers: http.Header{"Content-Type": {"image/gif"}}, body: []byte{40, 40}},
		{key: "5", headers: http.Header{"Content-Type": {"image/webp"}}, body: []byte{50, 50, 50}},
	}

	for _, item := range items {
		rw := httptest.NewRecorder()
		hit, err := cache.GetAndWriteTo(item.key, rw)
		require.NoError(t, err)
		require.False(t, hit)
		err = cache.Save(item.key, item.headers, item.body)
		require.NoError(t, err)
	}

	require.Equal(t, cCap, len(cache.items))

	for n, item := range items {
		rw := httptest.NewRecorder()
		hit, err := cache.GetAndWriteTo(item.key, rw)
		require.NoError(t, err)
		if n < len(items)-cCap {
			require.False(t, hit)
		} else {
			require.True(t, hit)
			res := rw.Result()
			res.Body.Close()

			for k, v := range item.headers {
				require.Equal(t, v[0], res.Header.Get(k))
			}
			require.Equal(t, item.body, rw.Body.Bytes())
		}
	}

	require.Equal(t, cCap, len(cache.items))
}
