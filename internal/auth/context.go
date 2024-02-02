package auth

import "context"

type tgUIDKey struct{}

func TGUID(ctx context.Context) int64 {
	uid, _ := ctx.Value(tgUIDKey{}).(int64) //nolint:errcheck // anyway have to check empty string
	return uid
}

func WithTGUID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, tgUIDKey{}, uid)
}

type identityIDKey struct{}

func IdentityID(ctx context.Context) string {
	e, _ := ctx.Value(identityIDKey{}).(string) //nolint:errcheck // anyway have to check empty string
	return e
}

func WithIdentityID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, identityIDKey{}, id)
}
