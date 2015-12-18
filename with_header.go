package cmhttp

import "net/http"

func WithHeader(name, value string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get(name) == "" {
				r.Header.Set(name, value)
			}
			return c.Do(r)
		})
	}
}
