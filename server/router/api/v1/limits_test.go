package v1

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

/*
capped is every field the protos put a length on, against the constant that states the
same number here.

Both are plain numbers on purpose. The proto declares the cap so the generated clients
and the published API reference carry it; limits.go states it so the code that refuses a
caller reads like code. Neither derives the other, because a number worked out at
startup from a registry lookup is a number nobody can read and a rename nobody finds
until the Instance will not start.

Every capped field is listed, not one per limit. The same text is capped in more than
one place — a label where an Item is made and again where one is changed, a name on the
four messages for setting up, joining, being added and editing your own profile — and
caps that disagree refuse a caller at one door and not the other.
*/
var capped = map[string]int{
	"CreateItemRequest.label":            LimitItemLabel,
	"UpdateItemRequest.label":            LimitItemLabel,
	"CreateItemRequest.quantity":         LimitItemQuantity,
	"UpdateItemRequest.quantity":         LimitItemQuantity,
	"UpdateItemRequest.note":             LimitItemNote,
	"CreateListRequest.name":             LimitListName,
	"RenameListRequest.name":             LimitListName,
	"AddMemberRequest.name":              LimitMemberName,
	"UpdateOwnProfileRequest.name":       LimitMemberName,
	"CompleteSetupRequest.name":          LimitMemberName,
	"RequestJoinRequest.name":            LimitMemberName,
	"AddMemberRequest.email":             LimitMemberEmail,
	"UpdateOwnProfileRequest.email":      LimitMemberEmail,
	"CompleteSetupRequest.email":         LimitMemberEmail,
	"RequestJoinRequest.email":           LimitMemberEmail,
	"CreateGroupRequest.name":            LimitGroupName,
	"CreateAccessTokenRequest.name":      LimitTokenName,
	"InstanceSettings.name":              LimitInstanceName,
	"CompleteSetupRequest.instance_name": LimitInstanceName,
	"RequestJoinRequest.message":         LimitJoinMessage,
}

// cappedField is what one proto field says about its length: the cap it declares, and
// the sentence above it that carries the number into the API reference.
type cappedField struct {
	limit   int
	inWords int
	saysSo  bool
}

/*
The protos and the constants say the same numbers.

Checked both ways round. A constant that disagrees with its field means the Instance
refuses at a length the reference does not state. A capped field missing from `capped`
means somebody added a limit and only half the doors know about it.

The sentence is checked too, because the OpenAPI generator carries comments into
`description` and carries the constraint nowhere. Without the sentence the published
reference states no limit at all, which is the gap this was written to close.
*/
func TestTheProtosAndTheConstantsAgree(t *testing.T) {
	declared := capsInProtos(t)
	if len(declared) < 20 {
		t.Fatalf("read %d capped fields out of the protos, fewer than there are", len(declared))
	}

	for field, stated := range capped {
		found, ok := declared[field]
		if !ok {
			t.Errorf("%s has a constant here and no cap in the protos", field)
			continue
		}
		if found.limit != stated {
			t.Errorf("%s is capped at %d in the proto and %d here", field, found.limit, stated)
		}
		if !found.saysSo {
			t.Errorf("%s is capped and says nothing about it, so the API reference will not either", field)
			continue
		}
		if found.inWords != found.limit {
			t.Errorf("%s is capped at %d and its comment says %d", field, found.limit, found.inWords)
		}
	}

	for field := range declared {
		if _, ok := capped[field]; !ok {
			t.Errorf("%s is capped in the protos and nothing here knows the number", field)
		}
	}
}

// declaresACap matches a field carrying a length, and saysACap the sentence above it.
var (
	declaresACap = regexp.MustCompile(`^\s*(?:optional )?\w+ (\w+) = \d+ \[\(buf\.validate\.field\)\.string\.max_len = (\d+)\];\s*$`)
	saysACap     = regexp.MustCompile(`^\s*// At most (\d+) characters\.\s*$`)
)

// capsInProtos reads every capped field out of the .proto files, by message and field.
func capsInProtos(t *testing.T) map[string]cappedField {
	t.Helper()

	paths, err := filepath.Glob("../../../../proto/nooks/api/v1/*.proto")
	if err != nil || len(paths) == 0 {
		t.Fatalf("found no protos to read: %v", err)
	}

	found := map[string]cappedField{}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		message := ""
		lines := strings.Split(string(body), "\n")
		for i, line := range lines {
			if name, ok := strings.CutPrefix(line, "message "); ok {
				message = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(name), "{"))
				continue
			}
			declares := declaresACap.FindStringSubmatch(line)
			if declares == nil {
				continue
			}

			one := cappedField{limit: number(t, declares[2])}
			if i > 0 {
				if says := saysACap.FindStringSubmatch(lines[i-1]); says != nil {
					one.saysSo, one.inWords = true, number(t, says[1])
				}
			}
			found[fmt.Sprintf("%s.%s", message, declares[1])] = one
		}
	}
	return found
}

func number(t *testing.T, said string) int {
	t.Helper()
	value, err := strconv.Atoi(said)
	if err != nil {
		t.Fatalf("%q is not a number", said)
	}
	return value
}
