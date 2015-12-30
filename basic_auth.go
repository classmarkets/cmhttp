package cmhttp

import "net/http"

// BasicAuth sets all request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted.
func BasicAuth(username, password string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			r.SetBasicAuth(username, password)
			return c.Do(r)
		})
	}
}
