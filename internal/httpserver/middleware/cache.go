package middleware

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/bardex/minipic/internal/app"
)

func NewCache(cache *app.LruCache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()
		rc := cache.GetOrCreateItem(key)

		if rc.IsExists() {
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
			err := rc.Save(result.Header, body)
			if err == nil {
				cache.PushFront(rc)
			} else {
				log.Println(err)
			}
		}

		for k, v := range result.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(result.StatusCode)
		if _, err := w.Write(body); err != nil {
			log.Println(err)
		}
	})
}
