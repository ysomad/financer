// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: telegram/v1/identity.proto

package telegramv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/ysomad/financer/internal/gen/proto/telegram/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_7_0

const (
	// IdentityServiceName is the fully-qualified name of the IdentityService service.
	IdentityServiceName = "telegram.v1.IdentityService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// IdentityServiceCreateIdentityProcedure is the fully-qualified name of the IdentityService's
	// CreateIdentity RPC.
	IdentityServiceCreateIdentityProcedure = "/telegram.v1.IdentityService/CreateIdentity"
	// IdentityServiceGetIdentityProcedure is the fully-qualified name of the IdentityService's
	// GetIdentity RPC.
	IdentityServiceGetIdentityProcedure = "/telegram.v1.IdentityService/GetIdentity"
	// IdentityServiceUpdateIdentityProcedure is the fully-qualified name of the IdentityService's
	// UpdateIdentity RPC.
	IdentityServiceUpdateIdentityProcedure = "/telegram.v1.IdentityService/UpdateIdentity"
)

// IdentityServiceClient is a client for the telegram.v1.IdentityService service.
type IdentityServiceClient interface {
	CreateIdentity(context.Context, *connect.Request[v1.CreateIdentityRequest]) (*connect.Response[v1.Identity], error)
	GetIdentity(context.Context, *connect.Request[v1.GetIdentityRequest]) (*connect.Response[v1.Identity], error)
	UpdateIdentity(context.Context, *connect.Request[v1.UpdateIdentityRequest]) (*connect.Response[v1.Identity], error)
}

// NewIdentityServiceClient constructs a client for the telegram.v1.IdentityService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewIdentityServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) IdentityServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &identityServiceClient{
		createIdentity: connect.NewClient[v1.CreateIdentityRequest, v1.Identity](
			httpClient,
			baseURL+IdentityServiceCreateIdentityProcedure,
			opts...,
		),
		getIdentity: connect.NewClient[v1.GetIdentityRequest, v1.Identity](
			httpClient,
			baseURL+IdentityServiceGetIdentityProcedure,
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		updateIdentity: connect.NewClient[v1.UpdateIdentityRequest, v1.Identity](
			httpClient,
			baseURL+IdentityServiceUpdateIdentityProcedure,
			opts...,
		),
	}
}

// identityServiceClient implements IdentityServiceClient.
type identityServiceClient struct {
	createIdentity *connect.Client[v1.CreateIdentityRequest, v1.Identity]
	getIdentity    *connect.Client[v1.GetIdentityRequest, v1.Identity]
	updateIdentity *connect.Client[v1.UpdateIdentityRequest, v1.Identity]
}

// CreateIdentity calls telegram.v1.IdentityService.CreateIdentity.
func (c *identityServiceClient) CreateIdentity(ctx context.Context, req *connect.Request[v1.CreateIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return c.createIdentity.CallUnary(ctx, req)
}

// GetIdentity calls telegram.v1.IdentityService.GetIdentity.
func (c *identityServiceClient) GetIdentity(ctx context.Context, req *connect.Request[v1.GetIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return c.getIdentity.CallUnary(ctx, req)
}

// UpdateIdentity calls telegram.v1.IdentityService.UpdateIdentity.
func (c *identityServiceClient) UpdateIdentity(ctx context.Context, req *connect.Request[v1.UpdateIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return c.updateIdentity.CallUnary(ctx, req)
}

// IdentityServiceHandler is an implementation of the telegram.v1.IdentityService service.
type IdentityServiceHandler interface {
	CreateIdentity(context.Context, *connect.Request[v1.CreateIdentityRequest]) (*connect.Response[v1.Identity], error)
	GetIdentity(context.Context, *connect.Request[v1.GetIdentityRequest]) (*connect.Response[v1.Identity], error)
	UpdateIdentity(context.Context, *connect.Request[v1.UpdateIdentityRequest]) (*connect.Response[v1.Identity], error)
}

// NewIdentityServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewIdentityServiceHandler(svc IdentityServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	identityServiceCreateIdentityHandler := connect.NewUnaryHandler(
		IdentityServiceCreateIdentityProcedure,
		svc.CreateIdentity,
		opts...,
	)
	identityServiceGetIdentityHandler := connect.NewUnaryHandler(
		IdentityServiceGetIdentityProcedure,
		svc.GetIdentity,
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	identityServiceUpdateIdentityHandler := connect.NewUnaryHandler(
		IdentityServiceUpdateIdentityProcedure,
		svc.UpdateIdentity,
		opts...,
	)
	return "/telegram.v1.IdentityService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case IdentityServiceCreateIdentityProcedure:
			identityServiceCreateIdentityHandler.ServeHTTP(w, r)
		case IdentityServiceGetIdentityProcedure:
			identityServiceGetIdentityHandler.ServeHTTP(w, r)
		case IdentityServiceUpdateIdentityProcedure:
			identityServiceUpdateIdentityHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedIdentityServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedIdentityServiceHandler struct{}

func (UnimplementedIdentityServiceHandler) CreateIdentity(context.Context, *connect.Request[v1.CreateIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("telegram.v1.IdentityService.CreateIdentity is not implemented"))
}

func (UnimplementedIdentityServiceHandler) GetIdentity(context.Context, *connect.Request[v1.GetIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("telegram.v1.IdentityService.GetIdentity is not implemented"))
}

func (UnimplementedIdentityServiceHandler) UpdateIdentity(context.Context, *connect.Request[v1.UpdateIdentityRequest]) (*connect.Response[v1.Identity], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("telegram.v1.IdentityService.UpdateIdentity is not implemented"))
}
