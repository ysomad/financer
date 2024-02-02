package rpc

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
)

// NewTelegramInterceptor returns interceptor which is checking if telegram api key matches.
func NewTelegramInterceptor(apiKey string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			hdr := req.Header().Get("X-API-KEY")
			if hdr == "" || hdr != apiKey {
				slog.Info("unauthenticated access", "api_key", hdr, "addr", req.Peer().Addr)
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
			}

			return next(ctx, req)
		})
	})
}
