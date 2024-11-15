package xfiber

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// XCacheRefresh is a header key to force refresh cache, the value is the cache key
const XCacheRefresh = "X-Cache-Refresh"

// XCacheHostname is used when XCacheRefresh is set, the value is the hostname
const XCacheHostname = "X-Cache-Hostname"

var ignoreHeaders = map[string]any{
	"Connection":          nil,
	"Keep-Alive":          nil,
	"Proxy-Authenticate":  nil,
	"Proxy-Authorization": nil,
	"TE":                  nil,
	"Trailers":            nil,
	"Transfer-Encoding":   nil,
	"Upgrade":             nil,
}

// cachedResponse will be saved in redis
type cachedResponse struct {
	Header map[string][]byte
	Body   []byte
	At     time.Time
}

// MarshalBinary implements encoding.BinaryMarshaler
func (v cachedResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (v *cachedResponse) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, v)
}

// Write cached resp to echo
func (v *cachedResponse) Write(c *fiber.Ctx) {
	c.Response().SetStatusCode(http.StatusOK)
	c.Response().SetBodyRaw(v.Body)
	for k, value := range v.Header {
		c.Response().Header.SetBytesV(k, value)
	}
}
