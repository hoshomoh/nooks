// Package password hashes and checks Member passwords, and owns the one rule nooks
// puts on them.
//
// The rule is deliberately singular: twelve characters or more, and nothing else. No
// character classes, no expiry, no strength meter. A household does not need a policy
// it will work around.
package password

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// MinLength is the only rule. It counts runes, not bytes, so a passphrase in any
// script is measured the way its author would count it.
const MinLength = 12

// MaxLength exists because bcrypt silently truncates beyond 72 bytes; a password
// longer than this would appear to work while ignoring its tail.
const MaxLength = 72

// ErrTooShort reports a password below MinLength.
var ErrTooShort = fmt.Errorf("a password needs %d characters or more", MinLength)

// ErrTooLong reports a password above what bcrypt can carry without truncating.
var ErrTooLong = fmt.Errorf("a password can be at most %d bytes", MaxLength)

// ErrWrong reports that a password does not match the stored hash. It is deliberately
// indistinguishable from "no such Member" at the call site, so sign-in cannot be used
// to discover who has an account.
var ErrWrong = errors.New("that password is not right")

// Validate reports whether a password may be used, without hashing it.
func Validate(plain string) error {
	if utf8.RuneCountInString(plain) < MinLength {
		return ErrTooShort
	}
	if len(plain) > MaxLength {
		return ErrTooLong
	}
	return nil
}

/*
Cost is the work bcrypt is asked for, and it is bcrypt's default: slow enough on the
modest hardware a household server tends to be, which is the whole point of the
algorithm.

A variable rather than a constant, for the suite and nothing else. Being slow on purpose
costs the same under the race detector, where everything is slower again, and the API
package alone hashes enough times to pass the ten minutes `go test` allows a package
before it gives up. That is how a commit touching no Go at all failed CI.

Tests lower it, in a TestMain, and `TestNothingButATestLowersTheCostOfAPassword` holds
every other file in the tree to leaving it alone. Lowered in a running Instance this
would be a real weakness, so it is guarded rather than trusted.
*/
var Cost = bcrypt.DefaultCost

// Hash validates a password and returns its hash.
func Hash(plain string) (string, error) {
	if err := Validate(plain); err != nil {
		return "", err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), Cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

// Verify checks a password against a stored hash, returning ErrWrong when it does not
// match. It does not apply Validate: a Member whose password predates a rule change
// must still be able to sign in and replace it.
func Verify(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrong
		}
		return fmt.Errorf("verify password: %w", err)
	}
	return nil
}
