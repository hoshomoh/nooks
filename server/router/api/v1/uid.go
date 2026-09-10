package v1

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

// uidBytes is the entropy in a public identifier. Ten bytes is 80 bits — far beyond
// collision in a household, and short enough to read out loud.
const uidBytes = 10

// uidEncoding is lower-case base32 without padding: unambiguous when written down, and
// safe in a URL without escaping.
var uidEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// newMemberUID makes a public identifier for a Member.
//
// It is random rather than sequential so that a resource name does not disclose how
// many Members an Instance has, or the order they joined in.
func newMemberUID() (string, error) {
	raw := make([]byte, uidBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return strings.ToLower(uidEncoding.EncodeToString(raw)), nil
}
