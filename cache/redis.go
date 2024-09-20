package cache

import (
	"context"
	"strconv"
)

const (
	RedisDefaultHost        = "localhost"
	RedisDefaultPort        = "6379"
	RedisSkipRequestedError = "redis-skipped-by-context"

	RedisForceRefreshKey = "RedisRefresh"
	RedisSkipKey         = "RedisSkip"

	RedisShortExpSeconds   = 60                //1 minute
	RedisDefaultExpSeconds = 60 * 10           //10 minutes
	RedisMediumExpSeconds  = 60 * 30           //30 minutes
	RedisLongExpSeconds    = 60 * 60           //1 hour
	RedisInfiniteExp       = 60 * 60 * 24 * 30 //1 month
	RedisOneDayExp         = 60 * 60 * 24      //1 day
	RedisVeryLongUserExp   = 60 * 60 * 2       //2 hours
	RedisNotAvailableError = "redis-not-available"

	RedisMsCKGetUserSkillXPs = "msg|usk|xp"
)

func GetContextWithSkipCache(ctx context.Context, val bool) context.Context {
	return context.WithValue(ctx, RedisSkipKey, val)
}

func IsForceRefreshCacheContext(ctx context.Context) bool {
	v, ok := ctx.Value(RedisForceRefreshKey).(bool)
	if !ok {
		return false
	}

	return v
}

func IsSkipCacheContext(ctx context.Context) bool {
	v, ok := ctx.Value(RedisSkipKey).(bool)
	if !ok {
		return false
	}

	return v
}

func GetContextWithForceRefreshCache(ctx context.Context, val bool) context.Context {
	return context.WithValue(ctx, RedisForceRefreshKey, val)
}

func RedisGetUserSkillXPsCacheKey(uid uint32, lang string) string {
	return RedisMsCKGetUserSkillXPs + "|" + strconv.Itoa(int(uid)) + "|" + lang
}
