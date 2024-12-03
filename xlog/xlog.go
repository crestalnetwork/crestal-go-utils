// Package xlog is a simple wrapper for slog, it will create a slog.Logger with options
package xlog

import (
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	slogslack "github.com/samber/slog-slack/v2"
)

// Options is the options for xlog, all options are optional
type Options struct {
	// Env will change the log output format, if Env is "local", log will output in text format, otherwise in json format
	Env string
	// Debug will change the log level to debug
	Debug bool
	// SlackToken is the Slack bot token, if it is set and SlackChannel exists, Warn and Error level will send to it
	SlackToken string
	// SlackChannel is the Slack channel id, if it is set and SlackToken exists, Warn and Error level will send to it
	SlackChannel string
	// Release is the release version of the service, it will be added to log fields
	Release string
	// ServiceName is the name of the service, it will be added to log fields
	ServiceName string
}

// New will create a new slog.Logger with options
func New(opts Options) *slog.Logger {
	var log *slog.Logger
	// init global variable log
	level := slog.LevelInfo
	if opts.Debug {
		level = slog.LevelDebug
	}
	if opts.Env == "local" {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	} else if opts.SlackToken != "" && opts.SlackChannel != "" {
		log = slog.New(
			slogmulti.Fanout(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
				slogslack.Option{
					Level:    slog.LevelWarn,
					BotToken: opts.SlackToken,
					Channel:  opts.SlackChannel,
				}.NewSlackHandler(),
			))
	} else {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	// add fields to log
	if opts.Env != "" {
		log = log.With("env", opts.Env)
	}
	if opts.Release != "" {
		log = log.With("release", opts.Release)
	}
	if opts.ServiceName != "" {
		log = log.With("service", opts.ServiceName)
	}
	slog.SetDefault(log) // some package use slog.Default() to get log, for example gorm

	return log
}
