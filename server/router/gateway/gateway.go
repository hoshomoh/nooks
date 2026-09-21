package gateway

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apiv1 "github.com/hoshomoh/nooks/proto/gen/nooks/api/v1"
	"github.com/hoshomoh/nooks/server/auth"
	v1 "github.com/hoshomoh/nooks/server/router/api/v1"
)

/*
New builds the REST mux.

The gateway does not run Connect's interceptors, so it needs its own way to work out
who is asking. It is the same function — resolver.Grant — called from a second place
rather than a second set of rules: the moment those two disagree is the moment a token
can do something through one door that it cannot through the other.
*/
func New(services v1.Services, resolver *auth.Resolver) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMiddlewares(attachGrant(resolver)),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
	)

	ctx := context.Background()
	register := []func() error{
		func() error {
			return apiv1.RegisterActivityServiceHandlerServer(ctx, mux, activityService{svc: services.Activity})
		},
		func() error { return apiv1.RegisterAuthServiceHandlerServer(ctx, mux, authService{svc: services.Auth}) },
		func() error {
			return apiv1.RegisterInstanceServiceHandlerServer(ctx, mux, instanceService{svc: services.Instance})
		},
		func() error { return apiv1.RegisterListServiceHandlerServer(ctx, mux, listService{svc: services.List}) },
		func() error {
			return apiv1.RegisterMemberServiceHandlerServer(ctx, mux, memberService{svc: services.Member})
		},
		func() error {
			return apiv1.RegisterPublicServiceHandlerServer(ctx, mux, publicService{svc: services.Public})
		},
		func() error {
			return apiv1.RegisterRequestServiceHandlerServer(ctx, mux, requestService{svc: services.Request})
		},
		func() error {
			return apiv1.RegisterTokenServiceHandlerServer(ctx, mux, tokenService{svc: services.Token})
		},
	}
	for _, one := range register {
		if err := one(); err != nil {
			return nil, err
		}
	}
	return mux, nil
}

/*
attachGrant puts the caller on the context, the same way the Connect interceptor does.

It never rejects. Whether a request is allowed is each service's decision, made against
the same Grant whichever door it came through — this only answers who is holding it.
*/
func attachGrant(resolver *auth.Resolver) runtime.Middleware {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			ctx := withHeaders(r.Context(), r.Header)
			if grant, ok := resolver.Grant(ctx, r.Header); ok {
				ctx = auth.WithGrant(ctx, grant)
			}
			next(w, r.WithContext(ctx), params)
		}
	}
}

// headerKey is unexported, so the only headers a service can be handed this way are the
// ones a request actually arrived with.
type headerKey struct{}

func withHeaders(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, headerKey{}, header)
}

/*
requestFrom builds the Connect request a service expects, carrying what the REST call
arrived with.

The services are written for Connect and read headers off the request. An adapter that
built an empty one left every such service seeing nothing: RefreshAccess reads the
refresh cookie, so the documented POST /api/v1/auth/refresh answered "not signed in" to
a caller holding a perfectly good one, and would have gone on doing that for any RPC
that started reading a header.
*/
func requestFrom[T any](ctx context.Context, msg *T) *connect.Request[T] {
	request := connect.NewRequest(msg)
	header, ok := ctx.Value(headerKey{}).(http.Header)
	if !ok {
		return request
	}
	for name, values := range header {
		for _, value := range values {
			request.Header().Add(name, value)
		}
	}
	return request
}

/*
asStatus turns a Connect error into the gRPC status the gateway knows how to render.

The codes are the same set with different names, so nothing is invented here: a service
that says permission_denied ends up as 403 whichever door the request came through.
*/
func asStatus(err error) error {
	var failure *connect.Error
	if !errors.As(err, &failure) {
		return status.Error(codes.Internal, err.Error())
	}
	return status.Error(codes.Code(failure.Code()), failure.Message())
}
