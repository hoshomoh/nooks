package v1

import (
	"fmt"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

/*
The most a Member may write into each field, read from the protos that declare it.

Nothing bounded any of these once. A request is capped at four mebibytes, so a List
could be named with four mebibytes of text and every screen drawing that name would try
to. The numbers are what somebody could mean rather than what a column could hold: a
long shopping line is under sixty characters and a long recipe is about four thousand.

They were constants here, and that made this the source. Everything else copied them:
the web kept a LIMITS object with a test holding it to this file, the MCP descriptions
were built from these, and the REST reference said nothing at all because the protos
carried no limits to generate from. Four surfaces, three of them copies, and adding a
field meant remembering all four.

The proto says it now, once, as a constraint the generators can read, and this reads it
back. The sentence a caller is refused with stays here, because "an item is longer than
500 characters" was written for whoever is typing and a generated validator's message
is about a field path and a rule id.

Counted in characters rather than bytes, which is what the number means to whoever is
typing. protovalidate's max_len counts Unicode code points, which is what
utf8.RuneCountInString counts, so the two agree rather than nearly agreeing.
*/
var (
	LimitItemLabel    = maxLenOf("nooks.api.v1.CreateItemRequest", "label")
	LimitItemQuantity = maxLenOf("nooks.api.v1.CreateItemRequest", "quantity")
	LimitItemNote     = maxLenOf("nooks.api.v1.UpdateItemRequest", "note")
	LimitListName     = maxLenOf("nooks.api.v1.CreateListRequest", "name")
	LimitMemberName   = maxLenOf("nooks.api.v1.AddMemberRequest", "name")
	LimitMemberEmail  = maxLenOf("nooks.api.v1.AddMemberRequest", "email")
	LimitGroupName    = maxLenOf("nooks.api.v1.CreateGroupRequest", "name")
	LimitTokenName    = maxLenOf("nooks.api.v1.CreateAccessTokenRequest", "name")
	LimitInstanceName = maxLenOf("nooks.api.v1.InstanceSettings", "name")
	LimitJoinMessage  = maxLenOf("nooks.api.v1.RequestJoinRequest", "message")
)

/*
maxLenOf is the cap one proto field declares.

It panics rather than returning an error, and it runs at package initialisation, so a
field that lost its constraint stops the Instance from starting instead of quietly
letting four mebibytes through. There is no useful way to carry on: the number is the
rule, and a zero here would be a rule that refuses everything.

The same limit is declared on more than one field — a label is capped on both creating
and updating an Item — and only one of them is read here. limits_test.go is what holds
the rest of them to it.
*/
func maxLenOf(message, field string) int {
	descriptor, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(message))
	if err != nil {
		panic(fmt.Sprintf("limits: no message %s: %v", message, err))
	}
	asMessage, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		panic(fmt.Sprintf("limits: %s is a %T, not a message", message, descriptor))
	}

	one := asMessage.Fields().ByName(protoreflect.Name(field))
	if one == nil {
		panic(fmt.Sprintf("limits: %s has no field %s", message, field))
	}

	rules, ok := proto.GetExtension(one.Options(), validate.E_Field).(*validate.FieldRules)
	if !ok || rules.GetString_() == nil || rules.GetString_().MaxLen == nil {
		panic(fmt.Sprintf("limits: %s.%s declares no max_len", message, field))
	}
	return int(rules.GetString_().GetMaxLen())
}
