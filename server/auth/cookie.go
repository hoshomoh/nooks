package auth

import (
	"net/http"
	"time"
)

// CookieName is the session cookie. It is prefixed so it cannot be set by a
// subdomain, and named for the app so a shared host stays legible.
const CookieName = "nooks_session"

// SessionLifetime is how long a session lasts without signing in again. A household
// device should not be asked every week.
const SessionLifetime = 30 * 24 * time.Hour

/*
AccessLifetime is how long the token handed over in a response body lasts.

An hour, because that is the credential a caller can actually lose: it travels in
bodies, gets pasted into terminals, and is held in memory by clients Nooks did not
write. The refresh token that mints it lives a month and never leaves the cookie.
*/
const AccessLifetime = time.Hour

// NewCookie builds the cookie carrying a session token.
//
// secure is passed in rather than sniffed, because whether the Instance is behind TLS
// is a deployment fact the server knows and this package does not.
func NewCookie(token string, expires time.Time, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:  CookieName,
		Value: token,
		Path:  "/",
		// HttpOnly: no script needs the token, and it stops an XSS bug becoming an
		// account takeover.
		HttpOnly: true,
		// Lax rather than Strict: a Member following a link to their own Public list
		// should still arrive signed in.
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  expires,
	}
}

// ExpiredCookie builds the cookie that clears a session.
func ExpiredCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	}
}
