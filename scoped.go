package cmhttp

import (
	"net/http"
	"net/url"
)

// Scoped resolves the Request's URL against the given baseURL before
// sending the request. If baseURL cannot be parsed, Scoped panics.
func Scoped(baseURL string) Decorator {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	return ScopedURL(u)
}

// ScopedURL resolves the Request's URL against the given baseURL before
// sending the request.
func ScopedURL(baseURL *url.URL) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			r.URL = baseURL.ResolveReference(r.URL)
			return c.Do(r)
		})
	}
}
