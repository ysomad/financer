package slogx

import (
	"context"
	"log/slog"
)

var _ slog.Handler = &handlerMiddleware{}

type handlerMiddleware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) *handlerMiddleware {
	return &handlerMiddleware{next: next}
}

func (h *handlerMiddleware) Enabled(ctx context.Context, l slog.Level) bool {
	return h.next.Enabled(ctx, l)
}

func (h *handlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.next.WithAttrs(attrs)
}

func (h *handlerMiddleware) WithGroup(name string) slog.Handler {
	return h.next.WithGroup(name)
}

func (h *handlerMiddleware) Handle(ctx context.Context, req slog.Record) error {
	if c, ok := ctx.Value(ctxKey{}).(logCtx); ok {
		if c.Recipient != "" {
			req.Add("recipient", c.Recipient)
		}
		if c.Version != "" {
			req.Add("version", c.Version)
		}
	}

	return h.next.Handle(ctx, req)
}
