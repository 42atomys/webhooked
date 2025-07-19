package contextutil

import "context"

type ContextKey uint8

const (
	webhookSpecCtxKey ContextKey = iota
	requestCtxKey
	storeCtxKey
)

func WithWebhookSpec(ctx context.Context, spec any) context.Context {
	return context.WithValue(ctx, webhookSpecCtxKey, spec)
}

func WebhookSpecFromContext[T any](ctx context.Context) (T, bool) {
	value, ok := ctx.Value(webhookSpecCtxKey).(T)
	return value, ok
}

func WithRequestCtx(ctx context.Context, rctx any) context.Context {
	return context.WithValue(ctx, requestCtxKey, rctx)
}

func RequestCtxFromContext[T any](ctx context.Context) (T, bool) {
	value, ok := ctx.Value(requestCtxKey).(T)
	return value, ok
}

func WithStore(ctx context.Context, store any) context.Context {
	return context.WithValue(ctx, storeCtxKey, store)
}

func StoreFromContext[T any](ctx context.Context) (T, bool) {
	value, ok := ctx.Value(storeCtxKey).(T)
	return value, ok
}
