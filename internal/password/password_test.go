package password

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateEnforcesOnlyLength(t *testing.T) {
	cases := []struct {
		name  string
		plain string
		want  error
	}{
		{"eleven characters", strings.Repeat("a", 11), ErrTooShort},
		{"exactly twelve", strings.Repeat("a", 12), nil},
		{"no character classes required", "aaaaaaaaaaaa", nil},
		{"a passphrase", "correct horse battery staple", nil},
		{"beyond bcrypt's limit", strings.Repeat("a", MaxLength+1), ErrTooLong},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Validate(tc.plain); !errors.Is(got, tc.want) {
				t.Errorf("Validate(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

// Twelve is counted in runes, so a passphrase in any script is measured the way its
// author would count it.
func TestValidateCountsRunesNotBytes(t *testing.T) {
	if err := Validate("паролькоторый"); err != nil {
		t.Errorf("Validate(13 Cyrillic runes) = %v, want nil", err)
	}
	if err := Validate("парольэтот"); !errors.Is(err, ErrTooShort) {
		t.Errorf("Validate(10 Cyrillic runes) = %v, want ErrTooShort", err)
	}
}

func TestHashAndVerifyRoundTrip(t *testing.T) {
	const plain = "the good beans from the market"

	hash, err := Hash(plain)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if hash == plain {
		t.Fatal("Hash returned the password itself")
	}
	if err := Verify(hash, plain); err != nil {
		t.Errorf("Verify with the right password = %v, want nil", err)
	}
	if err := Verify(hash, "something else entirely"); !errors.Is(err, ErrWrong) {
		t.Errorf("Verify with the wrong password = %v, want ErrWrong", err)
	}
}

func TestHashRejectsAnInvalidPassword(t *testing.T) {
	if _, err := Hash("short"); !errors.Is(err, ErrTooShort) {
		t.Errorf("Hash(short) = %v, want ErrTooShort", err)
	}
}

// Two Members with the same password must not share a hash, or the database would
// leak which accounts to attack together.
func TestHashIsSalted(t *testing.T) {
	const plain = "the good beans from the market"

	first, err := Hash(plain)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	second, err := Hash(plain)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if first == second {
		t.Error("hashing the same password twice gave the same hash, want a per-hash salt")
	}
}

// Verify does not apply Validate: a Member whose password predates a rule change must
// still be able to sign in and replace it.
func TestVerifyAcceptsAPasswordThatWouldNoLongerValidate(t *testing.T) {
	hash, err := Hash(strings.Repeat("a", MinLength))
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if err := Verify(hash, strings.Repeat("a", MinLength)); err != nil {
		t.Errorf("Verify = %v, want nil", err)
	}
}
