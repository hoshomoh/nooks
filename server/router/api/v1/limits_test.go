package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// sameCap groups the fields that mean the same thing, because a label is capped where
// an Item is made and again where one is changed. limits.go reads one of each; these
// are the rest, and they have to agree or a caller is refused at one door and not the
// other.
var sameCap = map[string][]string{
	"nooks.api.v1.CreateItemRequest.label":       {"nooks.api.v1.UpdateItemRequest.label"},
	"nooks.api.v1.CreateItemRequest.quantity":    {"nooks.api.v1.UpdateItemRequest.quantity"},
	"nooks.api.v1.CreateListRequest.name":        {"nooks.api.v1.RenameListRequest.name"},
	"nooks.api.v1.AddMemberRequest.name":         {"nooks.api.v1.UpdateOwnProfileRequest.name", "nooks.api.v1.CompleteSetupRequest.name", "nooks.api.v1.RequestJoinRequest.name"},
	"nooks.api.v1.AddMemberRequest.email":        {"nooks.api.v1.UpdateOwnProfileRequest.email", "nooks.api.v1.CompleteSetupRequest.email", "nooks.api.v1.RequestJoinRequest.email"},
	"nooks.api.v1.InstanceSettings.name":         {"nooks.api.v1.CompleteSetupRequest.instance_name"},
	"nooks.api.v1.CreateGroupRequest.name":       {},
	"nooks.api.v1.CreateAccessTokenRequest.name": {},
	"nooks.api.v1.RequestJoinRequest.message":    {},
	"nooks.api.v1.UpdateItemRequest.note":        {},
}

/*
One field means one cap, wherever it is written.

The same text is capped in more than one place: a label where an Item is made and again
where one is changed, a name on four messages between setting up, joining, being added
and editing your own profile. limits.go reads one field per limit, so nothing stops
another from drifting except this.

It is not a list of numbers. The numbers live in the protos; this says which fields are
the same field, which is a decision somebody makes rather than something a parser can
work out from two ints that happen to match.
*/
func TestTheSameFieldHasTheSameCapEverywhere(t *testing.T) {
	for leader, followers := range sameCap {
		want := capOf(t, leader)
		for _, follower := range followers {
			if got := capOf(t, follower); got != want {
				t.Errorf("%s caps at %d and %s at %d, which refuses a caller at one door and not the other",
					leader, want, follower, got)
			}
		}
	}
}

/*
Every capped field says its number in words as well.

The constraint is what refuses a caller and what the generated clients carry. It is not
what a person reads: the OpenAPI generator writes comments into `description` and does
not write the constraint anywhere, so the published API reference says how long a label
may be only because somebody wrote a sentence beside it.

Two statements of one number is how a number goes wrong, so the sentence is checked
against the constraint rather than trusted. Read out of the .proto files, because the
comment is not in the descriptor.
*/
func TestEveryCapIsAlsoWrittenInWords(t *testing.T) {
	said := capsSaidInProtos(t)
	if len(said) < 20 {
		t.Fatalf("read %d sentences out of the protos, fewer than the fields that carry one", len(said))
	}

	for field, inWords := range said {
		if declared := capOf(t, field); declared != inWords {
			t.Errorf("%s is capped at %d and its comment says %d", field, declared, inWords)
		}
	}

	// And the other way round: a constraint nobody wrote a sentence for is a limit the
	// reference does not state, which is the gap this was written to close.
	forEachCappedField(t, func(name string, _ int) {
		if _, ok := said[name]; !ok {
			t.Errorf("%s is capped and says nothing about it, so the API reference will not either", name)
		}
	})
}

// capOf is the max_len one field declares.
func capOf(t *testing.T, field string) int {
	t.Helper()

	at := strings.LastIndex(field, ".")
	value := maxLenOfOrZero(field[:at], field[at+1:])
	if value == 0 {
		t.Fatalf("%s declares no max_len", field)
	}
	return value
}

// maxLenOfOrZero is maxLenOf without the panic, so a test can report rather than stop.
func maxLenOfOrZero(message, field string) int {
	descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(message))
	if err != nil {
		return 0
	}
	asMessage, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return 0
	}
	one := asMessage.Fields().ByName(protoreflect.Name(field))
	if one == nil {
		return 0
	}
	rules, ok := proto.GetExtension(one.Options(), validate.E_Field).(*validate.FieldRules)
	if !ok || rules.GetString_() == nil || rules.GetString_().MaxLen == nil {
		return 0
	}
	return int(rules.GetString_().GetMaxLen())
}

// forEachCappedField visits every field in this API that declares a max_len.
func forEachCappedField(t *testing.T, visit func(name string, limit int)) {
	t.Helper()

	protoregistry.GlobalFiles.RangeFiles(func(file protoreflect.FileDescriptor) bool {
		if !strings.HasPrefix(string(file.Package()), "nooks.api.v1") {
			return true
		}
		messages := file.Messages()
		for i := range messages.Len() {
			message := messages.Get(i)
			fields := message.Fields()
			for j := range fields.Len() {
				field := fields.Get(j)
				name := fmt.Sprintf("%s.%s", message.FullName(), field.Name())
				if limit := maxLenOfOrZero(string(message.FullName()), string(field.Name())); limit != 0 {
					visit(name, limit)
				}
			}
		}
		return true
	})
}

// saysACap matches the sentence a capped field carries, which is the only place the
// number reaches whoever reads the API reference.
var saysACap = regexp.MustCompile(`^\s*// At most (\d+) characters\.\s*$`)

// declaresAField matches a field that carries a max_len, so the sentence above it can be
// attached to the right name.
var declaresAField = regexp.MustCompile(`^\s*(?:optional )?\w+ (\w+) = \d+ \[\(buf\.validate\.field\)\.string\.max_len = \d+\];\s*$`)

// capsSaidInProtos reads the sentence above every capped field, by full field name.
func capsSaidInProtos(t *testing.T) map[string]int {
	t.Helper()

	said := map[string]int{}
	paths, err := filepath.Glob("../../../../proto/nooks/api/v1/*.proto")
	if err != nil || len(paths) == 0 {
		t.Fatalf("found no protos to read: %v", err)
	}

	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		lines := strings.Split(string(body), "\n")
		message := ""
		for i, line := range lines {
			if strings.HasPrefix(line, "message ") {
				message = strings.TrimSuffix(strings.Fields(line)[1], "{")
				message = strings.TrimSpace(message)
			}
			field := declaresAField.FindStringSubmatch(line)
			if field == nil || i == 0 {
				continue
			}
			above := saysACap.FindStringSubmatch(lines[i-1])
			if above == nil {
				continue
			}
			limit, err := strconv.Atoi(above[1])
			if err != nil {
				t.Fatalf("%s: %q is not a number", path, above[1])
			}
			said[fmt.Sprintf("nooks.api.v1.%s.%s", message, field[1])] = limit
		}
	}
	return said
}
