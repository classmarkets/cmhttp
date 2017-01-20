package cmhttp

import (
	"net/http"
	"time"
)

// Logged is used to execute a log function after the request has been made.
// Neither the request body nor the response body will be logged.
// If the client returned an error it will be logged as well.
func Logged(logf func(string, ...interface{}), trigger func() bool) Decorator {
	if trigger == nil {
		trigger = func() bool { return true }
	}

	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (res *http.Response, err error) {
			if !trigger() {
				return c.Do(r)
			}

			defer func(begin time.Time) {
				took := time.Since(begin)

				if err != nil {
					logf(
						"cmhttp client error",
						"method", r.Method,
						"url", r.URL,
						"proto", r.Proto,
						"request_content_length", r.Header.Get("Content-Length"),
						"took_ms", int64(took/time.Millisecond),
						"error", err.Error(),
					)
				} else {
					logf(
						"cmhttp client response",
						"method", r.Method,
						"url", r.URL,
						"proto", r.Proto,
						"request_content_length", r.Header.Get("Content-Length"),
						"response_content_length", res.Header.Get("Content-Length"),
						"response_status", res.Status,
						"took_ms", int64(took/time.Millisecond),
					)
				}
			}(time.Now())

			res, err = c.Do(r)
			return res, err
		})
	}
}
