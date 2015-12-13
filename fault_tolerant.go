package cmhttp

import (
	"net/http"
	"time"
)

// FaultTolerant retries requests that failed do to network errors
// attempts times, and sleeps for increasing amounts of time between attempts.
// If all attempts fail, the last error is returned.
func FaultTolerant(attempts int, backoff time.Duration) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (res *http.Response, err error) {
			for i := 1; i <= attempts; i++ {
				if res, err = c.Do(r); err == nil {
					break
				}
				time.Sleep(backoff * time.Duration(i))
			}
			return res, err
		})
	}
}
