package redis

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/4books-sparta/utils/cache"
	"github.com/4books-sparta/utils/logging"

	"github.com/spf13/viper"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Enabled   bool
	IsCluster bool
	Host      string
	Nodes     string
	Port      string
	Password  string
	Database  int
	Timeout   time.Duration
}

func GetConfig() *Config {
	return &Config{
		Enabled:   viper.GetString("redis_cache_enabled") == "yes",
		Host:      viper.GetString("redis_host"),
		Port:      viper.GetString("redis_port"),
		Nodes:     viper.GetString("redis_nodes"),
		Password:  viper.GetString("redis_auth_token"),
		IsCluster: viper.GetString("redis_cluster") == "yes",
		Database:  0,
		Timeout:   500 * time.Millisecond,
	}
}

func (rc *Config) GetAddr() string {
	host := rc.Host
	if len(host) == 0 {
		host = cache.RedisDefaultHost
	}

	port := rc.Port
	if port == "" {
		port = cache.RedisDefaultPort
	}

	return host + ":" + port
}

func (rc *Config) GetAddresses() []string {
	if !rc.IsCluster {
		return []string{rc.GetAddr()}
	}
	return strings.Split(rc.Nodes, ",")
}

func RemoveKey(c redis.UniversalClient, key string) error {
	if c == nil {
		return errors.New(cache.RedisNotAvailableError)
	}

	return c.Del(context.Background(), key).Err()
}

func RemoveSlice(c redis.UniversalClient, keys []string) error {
	if c == nil {
		return errors.New(cache.RedisNotAvailableError)
	}

	return c.Del(context.Background(), keys...).Err()
}

func RemoveKeys(c redis.UniversalClient, startWith string) error {
	if c == nil {
		return errors.New(cache.RedisNotAvailableError)
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

func Store(c redis.UniversalClient, key string, value interface{}, seconds int, rep logging.ErrorReporter) error {
	if value == nil {
		//Skip
		return nil
	}
	if c == nil {
		fmt.Println("REDIS not available key: ", key)
		return errors.New(cache.RedisNotAvailableError)
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

func HStore(c redis.UniversalClient, key string, title string, value interface{}, seconds int, rep logging.ErrorReporter) error {
	if c == nil {
		fmt.Println("REDIS NOT AVAILABLE")
		return errors.New(cache.RedisNotAvailableError)
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

func SetKeyExpiration(c redis.UniversalClient, key string, seconds int) error {
	return c.Expire(context.Background(), key, time.Duration(seconds)*time.Second).Err()
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

func Get(ctx context.Context, c redis.UniversalClient, key string, dest interface{}) error {
	if c == nil {
		return errors.New(cache.RedisNotAvailableError)
	}
	if cache.IsForceRefreshCacheContext(ctx) || cache.IsSkipCacheContext(ctx) {
		return errors.New(cache.RedisSkipRequestedError)
	}

	p, err := c.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(p, dest)
}

func HGet(ctx context.Context, c redis.UniversalClient, key string, title string, dest interface{}) error {
	if c == nil {
		return errors.New(cache.RedisNotAvailableError)
	}
	if cache.IsForceRefreshCacheContext(ctx) || cache.IsSkipCacheContext(ctx) {
		return errors.New(cache.RedisSkipRequestedError)
	}

	p, err := c.HGet(context.Background(), key, title).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(p, dest)
}

func NewClient(config *Config) (redis.UniversalClient, error) {
	if !config.Enabled {
		fmt.Println("No-redis")
		return nil, errors.New("no-redis")
	}
	defOptions := redis.UniversalOptions{
		DialTimeout:  1 * config.Timeout,
		ReadTimeout:  1 * config.Timeout,
		WriteTimeout: 2 * config.Timeout,
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
