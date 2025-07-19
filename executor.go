package webhooked

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/42atomys/webhooked/internal/config"
	"github.com/42atomys/webhooked/internal/contextutil"
	"github.com/42atomys/webhooked/storage"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

type Executor interface {
	IncomingRequest(ctx context.Context, rctx *fasthttp.RequestCtx) error
}

type DefaultExecutor struct {
	workerPool sync.Pool
	wgPool     sync.Pool
}

type pipelineFn = func(ctx context.Context, rctx *fasthttp.RequestCtx, wh *config.Webhook) (context.Context, error)

func NewExecutor() *DefaultExecutor {
	return &DefaultExecutor{
		workerPool: sync.Pool{
			New: func() interface{} {
				slice := make([]byte, 0, 1024)
				return &slice
			},
		},
		wgPool: sync.Pool{
			New: func() interface{} {
				return &sync.WaitGroup{}
			},
		},
	}
}

func (e *DefaultExecutor) IncomingRequest(ctx context.Context, rctx *fasthttp.RequestCtx) error {
	wh, err := config.FetchWebhookByPath(rctx.Path())
	if errors.Is(err, config.ErrSpecNotFound) {
		return ErrHTTPNotFound(rctx, err)
	}
	log.Debug().Msgf("Resolved webhook spec: %v", wh.Name)

	ctx = contextutil.WithRequestCtx(ctx, rctx)
	ctx = contextutil.WithWebhookSpec(ctx, wh)

	for _, fn := range e.pipelineOrder() {
		if ctx, err = fn(ctx, rctx, wh); err != nil {
			return err
		}
	}

	return nil
}

func (e *DefaultExecutor) pipelineOrder() []pipelineFn {
	return []pipelineFn{
		e.pipelineSecure,
		e.pipelineStore,
		e.pipelineResponse,
	}
}

func (e *DefaultExecutor) pipelineSecure(ctx context.Context, rctx *fasthttp.RequestCtx, wh *config.Webhook) (context.Context, error) {
	if secure, err := wh.Security.IsSecure(rctx); err != nil || !secure {
		if err != nil {
			return ctx, ErrHTTPInternalServerError(rctx, fmt.Errorf("error during security validation: %w", err))
		}
		return ctx, ErrHTTPUnathorized(rctx, nil)
	}
	return ctx, nil
}

func (e *DefaultExecutor) pipelineStore(ctx context.Context, rctx *fasthttp.RequestCtx, wh *config.Webhook) (context.Context, error) {
	wg := e.wgPool.Get().(*sync.WaitGroup)
	defer e.wgPool.Put(wg)
	errChan := make(chan error)

	for _, store := range wh.Storage {
		storeCtx := contextutil.WithStore(ctx, store)
		wg.Add(1)

		go func(s *storage.Storage, gCtx context.Context) {
			payloadPtr := e.workerPool.Get().(*[]byte)
			payload := *payloadPtr

			defer func() {
				*payloadPtr = (*payloadPtr)[:0]
				e.workerPool.Put(payloadPtr)
				wg.Done()
			}()

			if s.Formatting.HasTemplate() {
				var err error
				payload, err = s.Formatting.Format(gCtx, map[string]any{})
				if err != nil {
					errChan <- err
					return
				}
			}

			if err := s.Store(gCtx, payload); err != nil {
				errChan <- err
				return
			}

		}(store, storeCtx)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return ctx, fmt.Errorf("error during the store of payload: %w", err)
		}
	}

	e.wgPool.Put(wg) // Put the WaitGroup back in the pool after all operations are complete
	return ctx, nil
}

func (e *DefaultExecutor) pipelineResponse(ctx context.Context, rctx *fasthttp.RequestCtx, wh *config.Webhook) (context.Context, error) {
	if !wh.Response.Formatting.HasTemplate() {
		rctx.SetStatusCode(fasthttp.StatusNoContent)
		return ctx, nil
	}

	response, err := wh.Response.Formatting.Format(ctx, map[string]any{})
	if err != nil {
		return ctx, ErrHTTPInternalServerError(rctx, fmt.Errorf("error formatting response: %w", err))
	}

	rctx.SetContentType(wh.Response.ContentType)
	rctx.SetStatusCode(wh.Response.StatusCode)
	rctx.SetBody(response)

	return ctx, nil
}
