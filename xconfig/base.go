package xconfig

import (
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogslack "github.com/samber/slog-slack/v2"
)

const (
	EnvLocal       = "local"
	EnvProduction  = "production" // this env only use in metric-related projects
	EnvTestnetDev  = "testnet-dev"
	EnvTestnetProd = "testnet-prod"
)

// Basic is the shared basic configuration for all projects
type Basic struct {
	Env     string `default:"local"`
	Debug   bool   `default:"false"`
	Release string `default:"local-debug"` // github build number, injected in image by github action
	// slack config is optional, if exists, it will send all warn/error log to slack
	SlackToken   string
	SlackChannel string `default:"C076H0HBZLZ"` // default is channel testnet-dev
}

// GenLogger generates a logger based on the environment and configuration
func (b Basic) GenLogger() *slog.Logger {
	var log *slog.Logger
	level := slog.LevelInfo
	if b.Debug {
		level = slog.LevelDebug
	}
	if b.Env == EnvLocal {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		}))
	} else if b.SlackToken != "" {
		log = slog.New(
			slogmulti.Fanout(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
				slogslack.Option{
					Level:    slog.LevelWarn,
					BotToken: b.SlackToken,
					Channel:  b.SlackChannel,
				}.NewSlackHandler(),
			))
	} else {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	return log.With("release", b.Release, "env", b.Env)
}
