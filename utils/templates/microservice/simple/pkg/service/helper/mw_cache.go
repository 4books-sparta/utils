package helper

import "github.com/go-redis/redis/v8"

type cacheMiddleware struct {
	noopMiddleware
	c redis.UniversalClient
}

func CachingMiddleware(client redis.UniversalClient) Middleware {
	return func(next Service) Service {
		return cacheMiddleware{
			noopMiddleware: noopMiddleware{next},
			c:              client,
		}
	}
}
