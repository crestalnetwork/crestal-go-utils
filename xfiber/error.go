// Package xfiber contains utilities for fiber framework
package xfiber

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/crestalnetwork/crestal-go-utils/xerr"
)

// ErrorHandler is a fiber error handler
func ErrorHandler(ctx *fiber.Ctx, err error) error {
	var final *xerr.Error

	// will check these types of errors
	var fe *fiber.Error
	var ve *validator.ValidationErrors

	if errors.As(err, &final) {
		// error already convert to final, will process it later
	} else if errors.As(err, &fe) {
		final = xerr.New(fe.Code, strings.ReplaceAll(http.StatusText(fe.Code), " ", ""), fe.Message)
	} else if errors.As(err, &ve) {
		final = xerr.Wrap(fiber.StatusBadRequest, "BadRequest", ve)
	} else if errors.Is(err, context.Canceled) {
		final = xerr.Wrap(fiber.StatusBadRequest, "ClientCancelled", err)
	} else {
		// other errors
		final = xerr.Wrap(fiber.StatusInternalServerError, "ServerError", err)
	}

	// log the internal server error
	if final.StatusCode() >= fiber.StatusInternalServerError {
		// log the error
		slog.Error("internal server error", "error", final, "component", "fiber")
	}

	return ctx.Status(final.StatusCode()).JSON(final)
}
