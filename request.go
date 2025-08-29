package webhooked

import (
	"fmt"

	"github.com/42atomys/webhooked/internal/fasthttpz"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

var (
	notFound            = []byte("Not Found")
	internalServerError = []byte("Internal Server Error")
	unauthorized        = []byte("Unauthorized")
	badRequest          = []byte("Bad Request")
)

func ErrHTTPNotFound(rctx *fasthttpz.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusNotFound)
	rctx.SetBody(notFound)
	if err != nil {
		return fmt.Errorf("not found: %w", err)
	}
	return nil
}

func ErrHTTPUnauthorized(rctx *fasthttpz.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusUnauthorized)
	rctx.SetBody(unauthorized)
	if err != nil {
		return fmt.Errorf("unauthorized: %w", err)
	}
	return nil
}

func ErrHTTPInternalServerError(rctx *fasthttpz.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusInternalServerError)
	rctx.SetBody(internalServerError)
	if err != nil {
		log.Error().Err(err).Msg(string(internalServerError))
		return fmt.Errorf("internal server error: %w", err)
	}
	return nil
}

func ErrHTTPBadRequest(rctx *fasthttpz.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusBadRequest)
	rctx.SetBody(badRequest)
	if err != nil {
		log.Error().Err(err).Msg(string(badRequest))
		return fmt.Errorf("bad request: %w", err)
	}
	return nil
}
