package slogx

import "context"

type ctxKey struct{}

type logCtx struct {
	Version   string
	Recipient string
}

func WithRecipient(ctx context.Context, recipient string) context.Context {
	if c, ok := ctx.Value(ctxKey{}).(logCtx); ok {
		c.Recipient = recipient
		return context.WithValue(ctx, ctxKey{}, c)
	}
	return context.WithValue(ctx, ctxKey{}, logCtx{Recipient: recipient})
}

func WithVersion(ctx context.Context, version string) context.Context {
	if c, ok := ctx.Value(ctxKey{}).(logCtx); ok {
		c.Version = version
		return context.WithValue(ctx, ctxKey{}, c)
	}
	return context.WithValue(ctx, ctxKey{}, logCtx{Recipient: version})
}
