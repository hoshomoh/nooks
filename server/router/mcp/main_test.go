package mcp

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/hoshomoh/nooks/internal/password"
)

/*
TestMain makes hashing cheap for this package.

bcrypt is slow on purpose, and this package hashes on nearly every test: signing in,
first run, adding a Member, replacing a password. At the real cost, under the race
detector, that took the package past the ten minutes `go test` allows before it gives
up, which is what had CI failing on commits that touched no Go at all.

Nothing here is testing bcrypt. `internal/password` is where the real cost is exercised.
*/
func TestMain(m *testing.M) {
	password.Cost = bcrypt.MinCost
	m.Run()
}
