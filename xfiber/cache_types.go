package xfiber

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vmihailenco/msgpack/v5"
)

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

func (v cachedResponse) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(v)
}

func (v *cachedResponse) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, v)
}

// Write cached resp to echo
func (v *cachedResponse) Write(c *fiber.Ctx) {
	c.Response().SetStatusCode(http.StatusOK)
	c.Response().SetBodyRaw(v.Body)
	for k, value := range v.Header {
		c.Response().Header.SetBytesV(k, value)
	}
}
