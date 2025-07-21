package webhooked

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

var (
	notFound            = []byte("Not Found")
	internalServerError = []byte("Internal Server Error")
	unauthorized        = []byte("Unauthorized")
	badRequest          = []byte("Bad Request")
)

func ErrHTTPNotFound(rctx *fasthttp.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusNotFound)
	rctx.SetBody(notFound)
	if err != nil {
		return fmt.Errorf("not found: %w", err)
	}
	return nil
}

func ErrHTTPUnauthorized(rctx *fasthttp.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusUnauthorized)
	rctx.SetBody(unauthorized)
	if err != nil {
		return fmt.Errorf("unauthorized: %w", err)
	}
	return nil
}

func ErrHTTPInternalServerError(rctx *fasthttp.RequestCtx, err error) error {
	log.Error().Err(err).Msg(string(internalServerError))
	rctx.SetStatusCode(fasthttp.StatusInternalServerError)
	rctx.SetBody(internalServerError)
	if err != nil {
		return fmt.Errorf("internal server error: %w", err)
	}
	return nil
}

func ErrHTTPBadRequest(rctx *fasthttp.RequestCtx, err error) error {
	log.Error().Err(err).Msg(string(badRequest))
	rctx.SetStatusCode(fasthttp.StatusBadRequest)
	rctx.SetBody(badRequest)
	if err != nil {
		return fmt.Errorf("bad request: %w", err)
	}
	return nil
}
