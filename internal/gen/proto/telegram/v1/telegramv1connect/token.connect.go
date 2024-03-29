// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: telegram/v1/token.proto

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
const _ = connect.IsAtLeastVersion0_1_0

const (
	// AccessTokenServiceName is the fully-qualified name of the AccessTokenService service.
	AccessTokenServiceName = "telegram.v1.AccessTokenService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// AccessTokenServiceIssueAccessTokenProcedure is the fully-qualified name of the
	// AccessTokenService's IssueAccessToken RPC.
	AccessTokenServiceIssueAccessTokenProcedure = "/telegram.v1.AccessTokenService/IssueAccessToken"
)

// AccessTokenServiceClient is a client for the telegram.v1.AccessTokenService service.
type AccessTokenServiceClient interface {
	// IssueAccessToken issues token for user with telegram id by telegram bot. X-API-KEY header must be provided.
	IssueAccessToken(context.Context, *connect.Request[v1.IssueAccessTokenRequest]) (*connect.Response[v1.IssueAccessTokenResponse], error)
}

// NewAccessTokenServiceClient constructs a client for the telegram.v1.AccessTokenService service.
// By default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped
// responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewAccessTokenServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) AccessTokenServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &accessTokenServiceClient{
		issueAccessToken: connect.NewClient[v1.IssueAccessTokenRequest, v1.IssueAccessTokenResponse](
			httpClient,
			baseURL+AccessTokenServiceIssueAccessTokenProcedure,
			opts...,
		),
	}
}

// accessTokenServiceClient implements AccessTokenServiceClient.
type accessTokenServiceClient struct {
	issueAccessToken *connect.Client[v1.IssueAccessTokenRequest, v1.IssueAccessTokenResponse]
}

// IssueAccessToken calls telegram.v1.AccessTokenService.IssueAccessToken.
func (c *accessTokenServiceClient) IssueAccessToken(ctx context.Context, req *connect.Request[v1.IssueAccessTokenRequest]) (*connect.Response[v1.IssueAccessTokenResponse], error) {
	return c.issueAccessToken.CallUnary(ctx, req)
}

// AccessTokenServiceHandler is an implementation of the telegram.v1.AccessTokenService service.
type AccessTokenServiceHandler interface {
	// IssueAccessToken issues token for user with telegram id by telegram bot. X-API-KEY header must be provided.
	IssueAccessToken(context.Context, *connect.Request[v1.IssueAccessTokenRequest]) (*connect.Response[v1.IssueAccessTokenResponse], error)
}

// NewAccessTokenServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewAccessTokenServiceHandler(svc AccessTokenServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	accessTokenServiceIssueAccessTokenHandler := connect.NewUnaryHandler(
		AccessTokenServiceIssueAccessTokenProcedure,
		svc.IssueAccessToken,
		opts...,
	)
	return "/telegram.v1.AccessTokenService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AccessTokenServiceIssueAccessTokenProcedure:
			accessTokenServiceIssueAccessTokenHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedAccessTokenServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedAccessTokenServiceHandler struct{}

func (UnimplementedAccessTokenServiceHandler) IssueAccessToken(context.Context, *connect.Request[v1.IssueAccessTokenRequest]) (*connect.Response[v1.IssueAccessTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("telegram.v1.AccessTokenService.IssueAccessToken is not implemented"))
}
