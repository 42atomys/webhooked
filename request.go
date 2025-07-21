package webhooked

import (
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
	return err
}

func ErrHTTPUnauthorized(rctx *fasthttp.RequestCtx, err error) error {
	rctx.SetStatusCode(fasthttp.StatusUnauthorized)
	rctx.SetBody(unauthorized)
	return err
}

func ErrHTTPInternalServerError(rctx *fasthttp.RequestCtx, err error) error {
	log.Error().Err(err).Msg(string(internalServerError))
	rctx.SetStatusCode(fasthttp.StatusInternalServerError)
	rctx.SetBody(internalServerError)
	return err
}

func ErrHTTPBadRequest(rctx *fasthttp.RequestCtx, err error) error {
	log.Error().Err(err).Msg(string(badRequest))
	rctx.SetStatusCode(fasthttp.StatusBadRequest)
	rctx.SetBody(badRequest)
	return err
}
