package auth

import (
	"net/http"
	"time"
)

/*
The session cookie, under the two names it can have.

`__Host-` tells a browser to refuse the cookie unless it came from this exact host over
HTTPS with no Domain set, which is what stops a hostile subdomain planting a session for
everybody on the parent domain. The prefix is only meaningful with Secure, and an
Instance on a LAN over plain HTTP is a deployment this supports, so the name follows the
setting rather than being fixed.

Two names is a cost, and it is paid where they are read: a browser sends the one it has,
so every read has to accept either. It is also why turning secure cookies on signs
everybody out once, which is a name change rather than anything going wrong.
*/
const (
	CookieName       = "nooks_session"
	SecureCookieName = "__Host-" + CookieName
)

// CookieNameFor is the name the cookie takes for this deployment.
func CookieNameFor(secure bool) string {
	if secure {
		return SecureCookieName
	}
	return CookieName
}

/*
SessionTokenFrom reads the session token out of whichever cookie a browser sent.

Either name, because which one it holds depends on when it last signed in rather than on
how the Instance is configured now. A browser that still has the plain cookie after
secure cookies were turned on is answered once and then given the prefixed one.
*/
func SessionTokenFrom(header http.Header) string {
	request := &http.Request{Header: header}
	for _, name := range []string{SecureCookieName, CookieName} {
		if cookie, err := request.Cookie(name); err == nil && cookie.Value != "" {
			return cookie.Value
		}
	}
	return ""
}

// SessionLifetime is how long a session lasts without signing in again. A household
// device should not be asked every week.
const SessionLifetime = 30 * 24 * time.Hour

/*
AccessLifetime is how long the token handed over in a response body lasts.

An hour, because that is the credential a caller can actually lose: it travels in
bodies, gets pasted into terminals, and is held in memory by clients nooks did not
write. The refresh token that mints it lives a month and never leaves the cookie.
*/
const AccessLifetime = time.Hour

// NewCookie builds the cookie carrying a session token.
//
// secure is passed in rather than sniffed, because whether the Instance is behind TLS
// is a deployment fact the server knows and this package does not.
func NewCookie(token string, expires time.Time, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:  CookieNameFor(secure),
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

/*
ExpiredCookies clear a session, under both names.

Both, because signing out has to clear whichever one the browser is holding, and after
secure cookies are turned on that may still be the plain one. Clearing only the current
name would leave the other in the browser, which is the shape of a cookie nobody can get
rid of.

The plain name is cleared without Secure, because a browser on plain HTTP refuses a
Secure cookie outright and would keep the one being cleared.
*/
func ExpiredCookies(secure bool) []*http.Cookie {
	expired := func(name string, markSecure bool) *http.Cookie {
		return &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   markSecure,
			MaxAge:   -1,
		}
	}
	if !secure {
		return []*http.Cookie{expired(CookieName, false)}
	}
	return []*http.Cookie{expired(SecureCookieName, true), expired(CookieName, false)}
}
