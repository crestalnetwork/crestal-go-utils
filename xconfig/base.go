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

type Basic struct {
	Env     string `default:"local"`
	Debug   bool   `default:"false"`
	Release string `default:"local-debug"` // github build number, injected in image by github action
	// slack config is optional, if exists, it will send all warn/error log to slack
	SlackToken   string
	SlackChannel string `default:"C076H0HBZLZ"` // default is channel testnet-dev
}

func (b Basic) GenLogger() *slog.Logger {
	level := slog.LevelInfo
	if b.Debug {
		level = slog.LevelDebug
	}
	if b.Env == EnvLocal {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		}))
	}
	if b.SlackToken != "" {
		return slog.New(
			slogmulti.Fanout(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
				slogslack.Option{
					Level:    slog.LevelWarn,
					BotToken: b.SlackToken,
					Channel:  b.SlackChannel,
				}.NewSlackHandler(),
			))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
