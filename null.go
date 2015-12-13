package cmhttp

import "net/http"

// Null always returns a 204 No Content response without ever opening a
// network connection.
func Null() Decorator {
	return func(c Client) Client {
		return ClientFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				Status:        http.StatusText(http.StatusNoContent),
				StatusCode:    http.StatusNoContent,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        make(map[string][]string),
				ContentLength: 0,
			}, nil
		})
	}
}
