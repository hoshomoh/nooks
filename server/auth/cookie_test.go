package auth

import (
	"net/http"
	"testing"
	"time"
)

/*
The cookie takes the __Host- prefix exactly when it can carry it.

`__Host-` tells a browser to refuse the cookie unless it came from this host over HTTPS
with no Domain, which is what stops a hostile subdomain planting a session for everybody
on the parent domain. It is only meaningful with Secure, and an Instance on a LAN over
plain HTTP is a deployment this supports, so the name follows the setting.

The conditions the prefix requires are asserted too. A browser silently ignores a
prefixed cookie that fails any of them, so getting Path or Domain wrong would leave
nobody able to sign in with nothing on screen to say why.
*/
func TestTheCookieIsPrefixedOnlyWhenItCanBe(t *testing.T) {
	expires := time.Date(2026, time.April, 1, 9, 0, 0, 0, time.UTC)

	plain := NewCookie("token", expires, false)
	if plain.Name != CookieName {
		t.Errorf("name = %q on plain HTTP, want %q", plain.Name, CookieName)
	}

	secure := NewCookie("token", expires, true)
	if secure.Name != SecureCookieName {
		t.Errorf("name = %q behind TLS, want %q", secure.Name, SecureCookieName)
	}
	if !secure.Secure {
		t.Error("the prefixed cookie is not Secure, which a browser refuses outright")
	}
	if secure.Path != "/" {
		t.Errorf("path = %q, want /: the prefix requires it", secure.Path)
	}
	if secure.Domain != "" {
		t.Errorf("domain = %q, want none: the prefix requires it", secure.Domain)
	}
}

/*
A session is read under whichever name the browser is holding.

Which one that is depends on when they last signed in rather than on how the Instance is
configured now, so an Instance that has just turned secure cookies on is still being
sent the plain one by everybody who was already signed in.
*/
func TestASessionIsReadUnderEitherName(t *testing.T) {
	for _, one := range []struct{ what, name string }{
		{"the plain cookie", CookieName},
		{"the prefixed cookie", SecureCookieName},
	} {
		t.Run(one.what, func(t *testing.T) {
			header := http.Header{"Cookie": []string{one.name + "=the-token"}}

			if got := SessionTokenFrom(header); got != "the-token" {
				t.Errorf("read %q from %s, want the token", got, one.name)
			}
		})
	}

	if got := SessionTokenFrom(http.Header{}); got != "" {
		t.Errorf("read %q from no cookie at all, want nothing", got)
	}
}

/*
Signing out clears both names.

After secure cookies are turned on, a browser may still be holding the plain one.
Clearing only the current name would leave the other behind, which is a session cookie
nobody can get rid of. The plain one is cleared without Secure, because a browser on
plain HTTP refuses a Secure cookie and would keep the very thing being cleared.
*/
func TestSigningOutClearsBothNames(t *testing.T) {
	cleared := ExpiredCookies(true)

	names := make(map[string]*http.Cookie, len(cleared))
	for _, cookie := range cleared {
		names[cookie.Name] = cookie
		if cookie.MaxAge >= 0 {
			t.Errorf("%s is not being cleared: MaxAge = %d", cookie.Name, cookie.MaxAge)
		}
	}

	for _, name := range []string{CookieName, SecureCookieName} {
		if _, ok := names[name]; !ok {
			t.Errorf("%s is left in the browser", name)
		}
	}
	if names[CookieName] != nil && names[CookieName].Secure {
		t.Error("the plain cookie is cleared with Secure, which a plain-HTTP browser refuses")
	}
}
