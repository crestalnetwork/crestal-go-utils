package xfiber

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp"
)

// Cache has several echo middlewares which can cache echo handler
type Cache struct {
	kv           redis.UniversalClient
	log          *slog.Logger
	disabled     bool
	asyncRefresh bool
	refreshing   sync.Map // if asyncRefresh, will store the refreshing key for mutex
	skipper      func(c *fiber.Ctx) bool
}

type CacheOptions struct {
	// Redis client is required
	Redis redis.UniversalClient
	// Logger is optional
	Logger *slog.Logger
	// AsyncRefresh will return the outdated cache and refresh it in the background when expired
	AsyncRefresh bool
	// Disabled with a bool expression
	Disabled bool
	// Skipper will skip the cache middleware
	Skipper func(c *fiber.Ctx) bool
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
	var c = &Cache{
		asyncRefresh: opts.AsyncRefresh,
		disabled:     opts.Disabled,
		skipper:      opts.Skipper,
	}
	if opts.Logger != nil {
		c.log = opts.Logger
	} else {
		c.log = slog.Default()
	}
	if opts.Redis == nil {
		msg := "Redis client is required when creating a cache middleware"
		c.log.Error(msg)
		time.Sleep(1 * time.Second)
		panic(msg)
	}
	c.kv = opts.Redis
	return c
}

func (cc *Cache) cacheResp(key string, resp *fiber.Response) {
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
	// fixed cache a long time, can be released if not used
	err := cc.kv.Set(context.Background(), key, cachedResponse{
		Header: header,
		Body:   utils.CopyBytes(resp.Body()),
		At:     time.Now(),
	}, time.Hour*24*7).Err()
	if err != nil {
		cc.log.Error("cache to redis failed", "key", key, "error", err)
		return
	}
	cc.log.Debug("cached", "key", key)
}

// Custom can cache entity by id
func (cc *Cache) Custom(kf KeyFunction, ef ExpFunction) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if cc.disabled {
			return c.Next()
		}
		if cc.skipper != nil && cc.skipper(c) {
			return c.Next()
		}
		// only cache GET method
		if c.Method() != fiber.MethodGet {
			return c.Next()
		}
		// pick up internal force cache
		forceKey := c.Get(XCacheRefresh)
		forceHostname := c.Get(XCacheHostname)
		if forceKey != "" && forceHostname != "" {
			cc.log.Info("force cache", "key", forceKey, "hostname", forceHostname)
			// run once at the same time
			cc.refreshing.Store(forceKey, struct{}{})
			// hack the hostname
			c.Request().Header.Set(fiber.HeaderXForwardedHost, forceHostname)
			// run real handler
			if err := c.Next(); err != nil {
				return err
			}
			// cache the resp
			cc.cacheResp(forceKey, c.Response())
			cc.refreshing.Delete(forceKey)
			cc.log.Debug("force cache done", "key", forceKey)
			return nil
		}
		// start normal process
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
			cc.log.Info("cache missed", "key", key)
		} else if err != nil {
			return err
		} else {
			// check if expired first
			expAt := cResp.At.Add(exp)
			if expAt.Before(time.Now()) && !cc.asyncRefresh {
				// expired, and not async refresh, go to real handler later
			} else {
				// return cached resp in this case
				// but if expired and async refresh is true, will refresh in the background
				if expAt.Before(time.Now()) && cc.asyncRefresh {
					// refresh in the background
					if _, ok := cc.refreshing.Load(key); !ok {
						local := new(fasthttp.URI)
						c.Request().URI().CopyTo(local)
						local.SetHost("localhost")
						agent := fiber.Get(local.String())
						agent.Set(XCacheRefresh, key)
						agent.Set(XCacheHostname, c.Hostname())
						go func() {
							_, _, errs := agent.Bytes()
							if len(errs) > 0 {
								cc.log.Error("async refresh failed", "key", key, "error", errs[0])
							}
						}()
						cc.log.Info("async refresh sent", "key", key)
					}
					// will continue to return the cached resp
				}
				// hit, return cached resp
				cResp.Write(c)
				// extra cache headers
				leftSec := int(expAt.Sub(time.Now()).Seconds())
				if leftSec < 0 {
					leftSec = 0
				}
				c.Set(fiber.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", leftSec))
				// skip real handler and other middlewares
				cc.log.Info("cache hit", "key", key, "left", leftSec)
				return nil
			}
		}

		// run real handler
		if err = c.Next(); err != nil {
			return err
		}

		// cache the resp
		cc.cacheResp(key, c.Response())
		return nil
	}
}

// Normal can cache endpoint in exp, can not clear manually in this mode.
// Use this in GET handler, and success resp code must be 200 StatusOK
func (cc *Cache) Normal(exp time.Duration) fiber.Handler {
	return cc.Custom(
		func(c *fiber.Ctx) (string, error) {
			return fmt.Sprintf("cache:%s?%s", c.Request().URI().Path(), c.Request().URI().QueryString()), nil
		},
		cc.Exp(exp),
	)
}

// NormalWithDomain can cache endpoint in exp, can not clear manually in this mode.
// Use this in GET handler, and success resp code must be 200 StatusOK
func (cc *Cache) NormalWithDomain(exp time.Duration) fiber.Handler {
	return cc.Custom(
		func(c *fiber.Ctx) (string, error) {
			return fmt.Sprintf("cache:%s:%s?%s", c.Hostname(), c.Request().URI().Path(), c.Request().URI().QueryString()), nil
		},
		cc.Exp(exp),
	)
}
