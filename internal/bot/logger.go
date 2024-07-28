package bot

import (
	"context"
	"log/slog"
)

var _ slog.Handler = &slogMiddleware{}

type slogMiddleware struct {
	next slog.Handler
}

func NewSlogMiddleware(next slog.Handler) *slogMiddleware {
	return &slogMiddleware{next: next}
}

func (m *slogMiddleware) Enabled(ctx context.Context, l slog.Level) bool {
	return m.next.Enabled(ctx, l)
}

func (m *slogMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m.next.WithAttrs(attrs)
}

func (m *slogMiddleware) WithGroup(name string) slog.Handler {
	return m.next.WithGroup(name)
}

func (m *slogMiddleware) Handle(ctx context.Context, req slog.Record) error {
	if c, ok := ctx.Value(logCtxKey{}).(logCtx); ok {
		if c.Recipient != "" {
			req.Add("recipient", c.Recipient)
		}
		if c.Version != "" {
			req.Add("version", c.Version)
		}
	}

	return m.next.Handle(ctx, req)
}

type logCtxKey struct{}

type logCtx struct {
	Version   string
	Recipient string
}

func withRecipient(ctx context.Context, recipient string) context.Context {
	if c, ok := ctx.Value(logCtxKey{}).(logCtx); ok {
		c.Recipient = recipient
		return context.WithValue(ctx, logCtxKey{}, c)
	}
	return context.WithValue(ctx, logCtxKey{}, logCtx{Recipient: recipient})
}

func withVersion(ctx context.Context, version string) context.Context {
	if c, ok := ctx.Value(logCtxKey{}).(logCtx); ok {
		c.Version = version
		return context.WithValue(ctx, logCtxKey{}, c)
	}
	return context.WithValue(ctx, logCtxKey{}, logCtx{Recipient: version})
}
