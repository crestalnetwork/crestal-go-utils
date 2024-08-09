package xfiber

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"github.com/crestalnetwork/crestal-go-utils/xerr"
)

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
		final = xerr.New(fiber.StatusBadRequest, "BadRequest", ve.Error())
	} else {
		// other errors
		final = xerr.New(fiber.StatusInternalServerError, "ServerError", err.Error())
	}

	// log the internal server error
	if final.StatusCode() >= fiber.StatusInternalServerError {
		// log the error
		log.Errorw("internal server error", "error", final)
	}

	return ctx.Status(final.StatusCode()).JSON(final)
}
