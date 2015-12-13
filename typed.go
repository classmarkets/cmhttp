package cmhttp

import "net/http"

// Typed adds Content-Type and Accept request headers to the given value,
// unless the respective header is non-empty.
func Typed(contentType string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Content-Type") == "" {
				r.Header.Set("Content-Type", contentType)
			}
			if r.Header.Get("Accept") == "" {
				r.Header.Set("Accept", contentType)
			}
			return c.Do(r)
		})
	}
}

// JSON sets the Content-Type and Accept request headers to
// "application/json" (unless the requests already has non-empty headers).
func JSON() Decorator {
	return Typed("application/json")
}
