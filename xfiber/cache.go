package xfiber

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/redis/go-redis/v9"
)

// Cache has several echo middlewares which can cache echo handler
type Cache struct {
	kv       redis.UniversalClient
	log      *slog.Logger
	disabled bool
}

type CacheOptions struct {
	// Redis client is required
	Redis redis.UniversalClient
	// Logger is optional
	Logger *slog.Logger
	// Disabled with a bool expression
	Disabled bool
}

// KeyFunction compute a redis key from echo context
// If the error is caused by the user, please return 400, if it is a system error, please wrap it with reason.
type KeyFunction func(c *fiber.Ctx) (string, error)

// ExpFunction compute a redis exp from echo context
// If the error is caused by the user, please return 400, if it is a system error, please wrap it with reason.
type ExpFunction func(c *fiber.Ctx) (time.Duration, error)

// Exp is a wrapper for fixed expired time
func (*Cache) Exp(exp time.Duration) ExpFunction {
	return func(c *fiber.Ctx) (time.Duration, error) {
		return exp, nil
	}
}

// NewCache create a cache client
func NewCache(opts CacheOptions) *Cache {
	var c = new(Cache)
	if opts.Logger != nil {
		c.log = opts.Logger
	} else {
		c.log = slog.Default()
	}
	if opts.Disabled {
		c.disabled = true
	}
	if opts.Redis == nil {
		msg := "Redis client is required when creating a cache middleware"
		c.log.Error(msg)
		time.Sleep(1 * time.Second)
		panic(msg)
	}
	return c
}

func (cc *Cache) cacheResp(key string, exp time.Duration, resp *fiber.Response) {
	if resp.StatusCode() != http.StatusOK {
		cc.log.Debug("bad resp status code, skip cache", "status", resp.StatusCode())
		return
	}
	header := make(map[string][]byte)
	resp.Header.VisitAll(
		func(key, value []byte) {
			// create real copy
			keyS := string(key)
			if _, ok := ignoreHeaders[keyS]; !ok {
				header[keyS] = utils.CopyBytes(value)
			}
		},
	)
	err := cc.kv.Set(context.Background(), key, cachedResponse{
		Header: header,
		Body:   utils.CopyBytes(resp.Body()),
		At:     time.Now(),
	}, exp).Err()
	if err != nil {
		cc.log.Error("cache to redis failed", "key", key, "error", err)
		return
	}
	cc.log.Debug("cached", "key", key)
}

// Entity can cache entity by id
func (cc *Cache) Entity(kf KeyFunction, ef ExpFunction) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cc.disabled {
			return c.Next()
		}
		// run before real handler
		key, err := kf(c)
		if err != nil {
			return err
		}
		exp, err := ef(c)
		if err != nil {
			return err
		}
		cResp := new(cachedResponse)
		err = cc.kv.Get(c.Context(), key).Scan(cResp)
		if errors.Is(err, redis.Nil) {
			// missed, continue
			cc.log.Debug("cache missed", "key", key)
		} else if err != nil {
			return err
		} else {
			// hit, return cached resp
			cResp.Write(c)
			// extra cache headers
			expAt := cResp.At.Add(exp)
			c.Set(fiber.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", int(expAt.Sub(time.Now()).Seconds())))
			// skip real handler and other middlewares
			cc.log.Debug("cache hit", "key", key)
			return nil
		}

		// run real handler
		if err = c.Next(); err != nil {
			return err
		}

		// cache the resp
		cc.cacheResp(key, exp, c.Response())
		return nil
	}
}