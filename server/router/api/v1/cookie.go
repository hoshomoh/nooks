package v1

import "net/http"

// cookieValue pulls one cookie out of a raw Cookie header. http.Request is borrowed
// purely for its parsing; Connect gives headers rather than a request.
func cookieValue(header, name string) string {
	if header == "" {
		return ""
	}
	request := &http.Request{Header: http.Header{"Cookie": []string{header}}}
	cookie, err := request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
