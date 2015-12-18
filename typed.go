package cmhttp

// Typed adds Content-Type and Accept request headers to the given value,
// unless the respective header is non-empty.
func Typed(contentType string) Decorator {
	return func(c Client) Client {
		c = WithHeader("Content-Type", contentType)(c)
		c = WithHeader("Accept", contentType)(c)
		return c
	}
}

// JSON sets the Content-Type and Accept request headers to
// "application/json" (unless the requests already has non-empty headers).
func JSON() Decorator {
	return Typed("application/json")
}
