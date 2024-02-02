package rpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

var errUnauthenticated = errors.New("unauthenticated")

// NewAPIKeyInterceptor returns interceptor which is checking if api key matches
func NewAPIKeyInterceptor(apiKey string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			hdr := req.Header().Get("X-API-KEY")
			if hdr == "" || hdr != apiKey {
				slog.Info("request without api key",
					"api_key", hdr,
					"addr", req.Peer().Addr,
					"endpoint", req.Spec().Procedure)
				return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
			}

			return next(ctx, req)
		})
	})
}

func NewAccessTokenInterceptor(secretKey string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			hdr := req.Header().Get("Authorization")
			if hdr == "" {
				slog.Info("request without Authorization header",
					"endpoint", req.Spec().Procedure,
					"addr", req.Peer().Addr)
				return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
			}

			hdrparts := strings.Split(hdr, " ")
			if len(hdrparts) != 2 || hdrparts[0] != "Bearer" {
				slog.Info("request with invalid Authorization header",
					"endpoint", req.Spec().Procedure,
					"addr", req.Peer().Addr)
				return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
			}

			token, err := jwt.Parse(hdrparts[1], func(t *jwt.Token) (interface{}, error) { return secretKey, nil })
			if err != nil {
				slog.Info("access token not parsed",
					"endpoint", req.Spec().Procedure,
					"addr", req.Peer().Addr,
					"err", err.Error())
				return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
			}

			if !token.Valid {
				slog.Error("request with invalid access token",
					"endpoint", req.Spec().Procedure,
					"addr", req.Peer().Addr)
				return nil, connect.NewError(connect.CodeUnauthenticated, errUnauthenticated)
			}

			return next(ctx, req)
		})
	})
}
