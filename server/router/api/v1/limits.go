package v1

/*
The most a Member may write into each field.

Nothing bounded any of these once. A request is capped at four mebibytes, so a List
could be named with four mebibytes of text and every screen drawing that name would try
to. The numbers are what somebody could mean rather than what a column could hold: a
long shopping line is under sixty characters and a long recipe is about four thousand.

Counted in characters rather than bytes, because that is what the number means to
whoever is typing. Storage is already bounded by the request cap.

The protos declare the same numbers as `max_len`, which is what the generated clients
carry and what puts them in the published API reference. These are not read off that at
runtime: doing so cost a registry lookup and a type assertion per number, failed at
startup rather than at compile time when a message was renamed, and turned ten obvious
constants into something nobody could read. `limits_test.go` holds the two together
instead, in both directions, which is how everything else in this repository is held.
*/
const (
	LimitItemLabel    = 500
	LimitItemQuantity = 50
	LimitItemNote     = 64_000
	LimitListName     = 200
	LimitMemberName   = 100
	// The longest an address can be, from RFC 5321. Not a judgement.
	LimitMemberEmail  = 254
	LimitGroupName    = 200
	LimitTokenName    = 200
	LimitInstanceName = 200
	LimitJoinMessage  = 1_000
)
