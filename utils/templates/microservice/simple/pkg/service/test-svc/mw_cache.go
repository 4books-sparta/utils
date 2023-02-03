package --service-name--

import (

	"github.com/4books-sparta/utils"
	"github.com/go-redis/redis/v8"

)

type cacheMiddleware struct {
	noopMiddleware
	c   redis.UniversalClient
	rep utils.ErrorReporter
}


func CachingMiddleware(client redis.UniversalClient, rep utils.ErrorReporter) Middleware {
	return func(next Service) Service {
		return cacheMiddleware{
			noopMiddleware: noopMiddleware{next},
			c:              client,
			rep:            rep,
		}
	}
}

