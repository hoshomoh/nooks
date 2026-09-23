package v1

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/hoshomoh/nooks/store/storetest"
)

/*
Nothing answers a caller who never said who they are, unless it is meant to.

parity_test.go makes adding an RPC a decision about whether an assistant may reach it.
This makes it a decision about whether a stranger may, which is the one that matters
more: a guard is a line inside a method body, and forgetting it is silent. The check is
that the method refuses, not that it contains a particular call, because several of them
guard through a helper and the point is the behaviour rather than the spelling.

An RPC absent from openToAnyone below fails until somebody says which it is. Being on
that list only means "not required to refuse": SignIn answers a stranger and still says
unauthenticated when the stranger has the password wrong, so there is nothing to assert
about a listed one beyond somebody having thought about it.
*/
var openToAnyone = map[string]string{
	// First run, getting in, and getting out again.
	"AuthService.CompleteSetup": "first run: there is nobody to be yet",
	"AuthService.SignIn":        "how a Member becomes one",
	"AuthService.SignOut":       "ending a session nobody has is not an error",
	"AuthService.RefreshAccess": "the refresh cookie is the credential, not a session",

	// Asking to join, and asking for a way back in. A stranger is the caller by
	// definition, and each is rate-limited and single-use in its own way.
	"AuthService.RequestJoin":           "a stranger asking for an account",
	"AuthService.GetJoinRequest":        "reading an invitation by its own secret",
	"AuthService.CompleteJoin":          "accepting one",
	"AuthService.RequestPasswordReset":  "somebody locked out",
	"AuthService.GetResetRequest":       "reading a reset by its own secret",
	"AuthService.CompletePasswordReset": "finishing one",

	// What an Instance shows the outside world.
	"InstanceService.GetInstance": "the sign-in page needs a name before anyone signs in",
	"PublicService.GetPublicList": "the whole point of publishing a List",
}

func TestNoRPCAnswersAStrangerByAccident(t *testing.T) {
	s := storetest.Fresh(t)

	now := func() time.Time { return testClock }
	services := Services{
		Activity: NewActivityService(s, nil),
		Auth:     NewAuthService(s, AuthServiceOptions{}),
		Instance: NewInstanceService(s),
		List:     NewListService(s, now, nil),
		Member:   NewMemberService(s, now, nil),
		Public:   NewPublicService(s),
		Request:  NewRequestService(s, now),
		Token:    NewTokenService(s, now, nil, nil),
	}

	for _, svc := range servicesIn(services) {
		for _, call := range rpcsOn(svc) {
			t.Run(call.name, func(t *testing.T) {
				// No grant on the context: this is a caller off the internet.
				err := call.refuse(t.Context())

				if _, open := openToAnyone[call.name]; open {
					return
				}
				if err == nil {
					t.Fatalf("%s answered a caller with no session", call.name)
				}
				if got := connect.CodeOf(err); got != connect.CodeUnauthenticated {
					t.Errorf("%s refused with %v, want unauthenticated", call.name, got)
				}
			})
		}
	}
}

// anonymousCall is one RPC, ready to be asked a question by nobody.
type anonymousCall struct {
	name string
	// refuse calls the method with an empty request and answers what came back.
	refuse func(ctx context.Context) error
}

// servicesIn reads the bundle as a list, so a service added to it is covered without
// this test being edited.
func servicesIn(services Services) []reflect.Value {
	value := reflect.ValueOf(services)
	out := make([]reflect.Value, 0, value.NumField())
	for i := range value.NumField() {
		out = append(out, value.Field(i))
	}
	return out
}

/*
rpcsOn finds the service methods that answer a Connect request.

By shape rather than by name: an RPC is a method taking a context and a *connect.Request
and returning a *connect.Response and an error. Anything else on the struct is a helper
and not a door.
*/
func rpcsOn(svc reflect.Value) []anonymousCall {
	serviceName := strings.TrimPrefix(svc.Type().String(), "*v1.")
	serviceName = strings.TrimPrefix(serviceName, "*")

	calls := make([]anonymousCall, 0, svc.NumMethod())
	for i := range svc.NumMethod() {
		method := svc.Type().Method(i)
		signature := method.Type
		if signature.NumIn() != 3 || signature.NumOut() != 2 {
			continue
		}
		request := signature.In(2)
		if request.Kind() != reflect.Pointer || !strings.HasPrefix(request.String(), "*connect.Request[") {
			continue
		}

		calls = append(calls, anonymousCall{
			name: serviceName + "." + method.Name,
			refuse: func(ctx context.Context) error {
				// connect.NewRequest(new(Req)), built for whichever message this takes.
				message := reflect.New(request.Elem().Field(0).Type.Elem())
				empty := reflect.New(request.Elem())
				empty.Elem().Field(0).Set(message)

				out := svc.Method(i).Call([]reflect.Value{reflect.ValueOf(ctx), empty})
				if err, _ := out[1].Interface().(error); err != nil {
					return err
				}
				return nil
			},
		})
	}
	return calls
}
