package xfiber

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MidParseEnvStage is a middleware that parses the environment stage by the hostname in raw request context
// 1: local and testnet-dev
// 2: testnet-prod
func MidParseEnvStage(ctx *fiber.Ctx) error {
	var stage = 2
	hostname := ctx.Hostname()
	// remove port
	hostname = strings.Split(hostname, ":")[0]
	// check
	if !strings.Contains(hostname, ".") {
		// If the hostname does not contain a dot, it's a local development.
		stage = 1
	} else if strings.HasSuffix(hostname, ".crestal.dev") {
		// If the hostname ends with .crestal.dev, it's a staging environment.
		stage = 1
	} else if hostname == "127.0.0.1" {
		stage = 1
	}
	// otherwise, it's a production environment.
	ctx.Context().SetUserValue("stage", stage)
	return ctx.Next()
}

// EnvStage returns the environment stage by reading the value from the context
// 1: local and testnet-dev
// 2: testnet-prod
func EnvStage(ctx context.Context) int {
	stage := ctx.Value("stage")
	if stage == nil {
		return 2
	}
	v, ok := stage.(int)
	if !ok {
		return 2
	}
	return v
}
