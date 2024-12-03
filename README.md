# Crestal Go Utils

This package provides a toolbox for crestal golang projects.  
In these utilities, we prioritize the [12-factor](https://12factor.net/) principles.

[![go report card](https://goreportcard.com/badge/github.com/crestalnetwork/crestal-go-utils "go report card")](https://goreportcard.com/report/github.com/crestalnetwork/crestal-go-utils)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/crestalnetwork/crestal-go-utils?tab=doc)

## Logger
this package use standard slog package for logging, 
so if you set the global default logger anywhere, it will be used by this package.
```go
  slog.SetDefault(YOUR_LOGGER)
```

## xlog
A wrapper around the standard slog package that offers a New function for easily creating a logger.

## xconfig
Load configuration from environment variables, docker/k8s secrets, aws systems manager or secret manager.

## xerr
A custom error type which implements the error interface and provides additional information about the error.

## xfiber
Utilities for fiber web framework.

## xutils
Utilities for general purpose.
