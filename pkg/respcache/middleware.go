package respcache

import (
	"log"
	"net/http"
)

func NewCacheMiddleware(cache *LruCache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit, err := cache.Read(r.RequestURI, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if hit {
			return
		}

		rc := cache.CreateResponseCache(r.RequestURI)
		defer rc.Close()

		rc.WriteHeader(200)
		rc.SetHeader(w.Header().Clone())

		next.ServeHTTP(rc, r)

		if err := rc.Read(w); err != nil {
			log.Println(err)
			return
		}

		if rc.GetStatus() == 200 {
			// save response cache to lru collection
			cache.PushFront(rc)
		} else {
			// remove response cache
			rc.Remove()
		}
	})
}
