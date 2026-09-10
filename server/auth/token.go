// Package auth turns a session cookie into a Member, and back.
//
// It owns three things and nothing else: minting session tokens, carrying the signed-in
// Member on a context, and the Connect interceptor that joins the two. Password rules
// live in internal/password; persistence lives in store.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// tokenBytes is the entropy in a session token. 32 bytes is well beyond guessing and
// still fits comfortably in a cookie.
const tokenBytes = 32

// NewToken mints a session token and the hash to store for it.
//
// Only the hash is ever persisted, so a stolen database does not hand over live
// sessions. The token itself is returned once, to be put in the cookie.
func NewToken() (token string, hash string, err error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("read random bytes: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, HashToken(token), nil
}

// HashToken derives the stored form of a token.
//
// SHA-256 rather than bcrypt on purpose: a session token is 32 random bytes, so it has
// no guessable structure to slow an attacker down over, and this runs on every request.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
