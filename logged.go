package cmhttp

import (
	"net/http"
	"time"
)

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
						"took_ms", took/1e6,
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
						"took_ms", took/1e6,
					)
				}
			}(time.Now())

			res, err = c.Do(r)
			return res, err
		})
	}
}
