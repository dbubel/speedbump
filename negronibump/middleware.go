package negronibump

import (
	"net"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/etcinit/speedbump"
	"github.com/go-redis/redis"
	"github.com/unrolled/render"
)

func RateLimit(client *redis.Client, hasher speedbump.RateHasher, max int64) negroni.HandlerFunc {
	limiter := speedbump.NewLimiter(client, hasher, max)
	rnd := render.New()

	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		ok, err := limiter.Attempt(ip)
		if err != nil {
			panic(err)
		}

		if !ok {
			// nextTime := time.Now().Add(hasher.Duration())
			rnd.JSON(rw, 429, map[string]string{"error": "Rate limit exceeded. Try again in "})
		} else {
			next(rw, r)
		}
	}
}
