package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/bardex/minipic/internal/app"
)

func NewCache(cache *app.LruCache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()
		rc := cache.GetItem(key)

		if rc.NotEmpty() {
			if err := rc.WriteTo(w); err == nil {
				return
			}
		}

		rec := httptest.NewRecorder()

		next.ServeHTTP(rec, r)

		result := rec.Result()
		defer result.Body.Close()

		body, _ := ioutil.ReadAll(result.Body)

		if result.StatusCode == 200 {
			rc.Save(result.Header, body)
			cache.PushFront(rc)
		}

		for k, v := range result.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(result.StatusCode)
		w.Write(body)
	})
}
