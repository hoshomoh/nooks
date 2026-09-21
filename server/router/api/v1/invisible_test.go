package v1

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*
A List somebody cannot reach must answer exactly as a List that was never there.

This is what accessTo states: AccessNone means invisible, and invisible must not be
distinguishable from absent, or every endpoint taking a uid becomes a way to ask "does
this exist?" about somebody else's Lists. Five endpoints were checked by hand. Fourteen
take one.

Asserted as "the same answer" rather than as "not_found", because what an RPC does with
a uid it cannot find is its own business — it may want an argument first. What it may
not do is answer differently depending on whether the List is real, and that is a
comparison rather than a code.
*/
func TestAnInvisibleListAnswersLikeOneThatIsNotThere(t *testing.T) {
	f := newListFixture(t)

	// Jonas's, private, with something on it. Anna is an Admin and still cannot see it:
	// an Admin runs the Instance, they do not get to read everybody's Lists.
	hidden := f.createList(t, f.jonas, "Doctor")
	hiddenItem := f.addItem(t, f.jonas, hidden, "Blood test results")

	as := f.as(t, f.anna)
	svc := reflect.ValueOf(f.svc)
	addressed := 0

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
		message := request.Elem().Field(0).Type.Elem()

		naming, holds := uidFieldOf(message)
		if !holds {
			continue
		}
		addressed++

		ask := func(uid string) string {
			built := reflect.New(request.Elem())
			built.Elem().Field(0).Set(addressedTo(message, naming, uid))

			out := svc.Method(i).Call([]reflect.Value{reflect.ValueOf(as), built})
			if err, _ := out[1].Interface().(error); err != nil {
				return fmt.Sprintf("%v: %s", connect.CodeOf(err), err)
			}
			return "answered"
		}

		t.Run(method.Name, func(t *testing.T) {
			invisible := ask(uidFor(naming, hidden, hiddenItem))
			absent := ask(uidFor(naming, "lst_no_such_list", "itm_no_such_item"))

			if absent == "answered" {
				t.Fatalf("%s answered for a %s that does not exist", method.Name, naming)
			}
			if invisible != absent {
				t.Errorf("%s answers %q for somebody else's List and %q for one that is not there",
					method.Name, invisible, absent)
			}
		})
	}

	// 14 today. Stated as a floor so that adding an RPC does not need this edited, and
	// deleting the reflection above does not leave a test that passes over nothing.
	if addressed < 14 {
		t.Fatalf("only %d RPCs were addressed by uid, so this is checking less than it did", addressed)
	}
}

// uidFieldOf names the field an RPC addresses a List by, when it has one.
func uidFieldOf(message reflect.Type) (protoreflect.Name, bool) {
	fields := reflect.New(message).Interface().(proto.Message).ProtoReflect().Descriptor().Fields()
	for _, naming := range []protoreflect.Name{"list_uid", "item_uid"} {
		if fields.ByName(naming) != nil {
			return naming, true
		}
	}
	return "", false
}

// uidFor picks which of the two identifiers a field wants.
func uidFor(naming protoreflect.Name, listUID, itemUID string) string {
	if naming == "list_uid" {
		return listUID
	}
	return itemUID
}

// addressedTo builds the request message with the uid filled in and nothing else.
func addressedTo(message reflect.Type, naming protoreflect.Name, uid string) reflect.Value {
	built := reflect.New(message)
	filled := built.Interface().(proto.Message).ProtoReflect()
	filled.Set(filled.Descriptor().Fields().ByName(naming), protoreflect.ValueOfString(uid))
	return built
}
