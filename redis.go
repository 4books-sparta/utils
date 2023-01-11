package utils

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/go-redis/redis/v8"
)

const (
	RedisDefaultHost        = "localhost"
	RedisDefaultPort        = "6379"
	RedisSkipRequestedError = "redis-skipped-by-context"
	RedisForceRefreshKey    = "RedisRefresh"
	RedisSkipKey            = "RedisSkip"

	RedisShortExpSeconds   = 60                //1 minute
	RedisDefaultExpSeconds = 60 * 10           //10 minutes
	RedisMediumExpSeconds  = 60 * 30           //30 minutes
	RedisLongExpSeconds    = 60 * 60           //1 hour
	RedisInfiniteExp       = 60 * 60 * 24 * 30 //1 month
	RedisOneDayExp         = 60 * 60 * 24      //1 day
	RedisVeryLongUserExp   = 60 * 60 * 2       //2 hours
	RedisNotAvailableError = "redis-not-available"
)

type RedisConfig struct {
	Enabled   bool
	IsCluster bool
	Host      string
	Nodes     string
	Port      string
	Password  string
	Database  int
}

func GetRedisConfig() *RedisConfig {
	return &RedisConfig{
		Enabled:   viper.GetString("redis_cache_enabled") == "yes",
		Host:      viper.GetString("redis_host"),
		Port:      viper.GetString("redis_port"),
		Nodes:     viper.GetString("redis_nodes"),
		Password:  viper.GetString("redis_auth_token"),
		IsCluster: viper.GetString("redis_cluster") == "yes",
		Database:  0,
	}
}

func (rc *RedisConfig) GetAddr() string {
	host := rc.Host
	if len(host) == 0 {
		host = RedisDefaultHost
	}

	port := rc.Port
	if port == "" {
		port = RedisDefaultPort
	}

	return host + ":" + port
}

func (rc *RedisConfig) GetAddresses() []string {
	if !rc.IsCluster {
		return []string{rc.GetAddr()}
	}
	return strings.Split(rc.Nodes, ",")
}

func RedisRemoveKey(c redis.UniversalClient, key string) error {
	if c == nil {
		return errors.New(RedisNotAvailableError)
	}

	return c.Del(context.Background(), key).Err()
}

func RedisRemoveSlice(c redis.UniversalClient, keys []string) error {
	if c == nil {
		return errors.New(RedisNotAvailableError)
	}

	return c.Del(context.Background(), keys...).Err()
}

func RedisRemoveKeys(c redis.UniversalClient, startWith string) error {
	if c == nil {
		return errors.New(RedisNotAvailableError)
	}

	ctx := context.Background()
	iter := c.Scan(ctx, 0, startWith+"*", 200).Iterator()
	for iter.Next(ctx) {
		err := c.Del(ctx, iter.Val()).Err()
		if err != nil {
			fmt.Printf("RedisRemoveKeys Del report an error: %+v\n", err)
			return err
		}
	}
	if err := iter.Err(); err != nil {
		fmt.Printf("RedisRemoveKeys iter report an error: %+v\n", err)
		return err
	}

	return nil
}

func RedisStore(c redis.UniversalClient, key string, value interface{}, seconds int, rep ErrorReporter) error {
	if value == nil {
		//Skip
		return nil
	}
	if c == nil {
		fmt.Println("REDIS not available key: ", key)
		return errors.New(RedisNotAvailableError)
	}

	p, err := json.Marshal(value)
	if err != nil {
		fmt.Println("JSON marshal error: ", key)
		return err
	}

	go func() {
		err := c.Set(context.Background(), key, string(p), time.Duration(seconds)*time.Second).Err()
		if err != nil {
			fmt.Println("REDIS store error: ", key)
			//Unmanaged error
			rep.Report(err, "redis-store-error", "key", key)
		}
	}()

	return nil
}

func RedisHStore(c redis.UniversalClient, key string, title string, value interface{}, seconds int, rep ErrorReporter) error {
	if c == nil {
		fmt.Println("REDIS NOT AVAILABLE")
		return errors.New(RedisNotAvailableError)
	}

	val, err := SerializeValue(value)
	if err != nil {
		return err
	}

	go func(vv string) {
		ctx := context.Background()
		err := c.HSet(ctx, key, title, vv).Err()
		if err != nil {
			rep.Report(err, "key", key, "title", title, "val", vv)
		}

		if seconds > 0 {
			err = c.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
			if err != nil {
				rep.Report(err, "key", key, "title", title, "val", "expiration")
			}
		}
	}(val)

	return nil
}

func HashKey(i interface{}) (string, error) {
	data, err := json.Marshal(&i)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func SerializeValue(i interface{}) (string, error) {
	data, err := json.Marshal(&i)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func RedisGet(ctx context.Context, c redis.UniversalClient, key string, dest interface{}) error {
	if c == nil {
		return errors.New(RedisNotAvailableError)
	}
	if IsForceRefreshCacheContext(ctx) || IsSkipCacheContext(ctx) {
		return errors.New(RedisSkipRequestedError)
	}

	p, err := c.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(p, dest)
}

func RedisHGet(ctx context.Context, c redis.UniversalClient, key string, title string, dest interface{}) error {
	if c == nil {
		return errors.New(RedisNotAvailableError)
	}
	if IsForceRefreshCacheContext(ctx) || IsSkipCacheContext(ctx) {
		return errors.New(RedisSkipRequestedError)
	}

	p, err := c.HGet(context.Background(), key, title).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(p, dest)
}

func GetContextWithForceRefreshCache(ctx context.Context, val bool) context.Context {
	return context.WithValue(ctx, RedisForceRefreshKey, val)
}

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

func NewRedisClient(config *RedisConfig) (redis.UniversalClient, error) {
	if !config.Enabled {
		fmt.Println("No-redis")
		return nil, errors.New("no-redis")
	}
	defOptions := redis.UniversalOptions{
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	if !config.IsCluster {
		fmt.Println("- REDIS STANDALONE")
		return redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        []string{config.GetAddr()},
			DB:           config.Database,
			Password:     config.Password,
			DialTimeout:  defOptions.DialTimeout,
			ReadTimeout:  defOptions.ReadTimeout,
			WriteTimeout: defOptions.WriteTimeout,
		}), nil
	}

	addr := config.GetAddresses()
	if len(addr) <= 1 {
		//It's not a cluster
		//return a standalone redis
		fmt.Println("- REDIS STANDALONE FALLBACK")
		return redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        []string{config.GetAddr()},
			DB:           config.Database,
			Password:     config.Password,
			DialTimeout:  defOptions.DialTimeout,
			ReadTimeout:  defOptions.ReadTimeout,
			WriteTimeout: defOptions.WriteTimeout,
		}), nil
	}
	//Get The Master Name

	fmt.Println("- REDIS CLUSTER OK")
	return redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        addr,
		Password:     config.Password, // no password set
		DB:           config.Database, // use default DB
		DialTimeout:  defOptions.DialTimeout,
		ReadTimeout:  defOptions.ReadTimeout,
		WriteTimeout: defOptions.WriteTimeout,
	}), nil
}
