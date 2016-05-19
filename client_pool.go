package cmhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/bitly/go-hostpool"
)

// StaticClientPool creates a pool of HTTP clients that use an ε-greedy strategy to distribute
// HTTP requests among multiple hosts. The pool transparently distributes the requests
// among the hosts taking the individual request durations and failures into account.
//
// The first parameter must be a list of absolute URLs that correspond to the hosts
// that should receive the client requests. If any of the given URLs is not valid or
// relative StaticClientPool will panic immediately instead of when the returned decorator
// is actually used.
//
// A more detailed discussion of the underlying algorithm can be found at
// https://godoc.org/github.com/bitly/go-hostpool#NewEpsilonGreedy
//
// See also https://en.wikipedia.org/wiki/Epsilon-greedy_strategy
func StaticClientPool(baseURLs []string, decayDuration time.Duration, valueCalculator hostpool.EpsilonValueCalculator) Decorator {
	for i := range baseURLs {
		// check for each host if we can actually parse the URL so we can
		// fail immediately when creating this decorator instead of
		// waiting until it is used later.
		u, err := url.Parse(baseURLs[i])
		if err != nil {
			panic(err)
		}

		if !u.IsAbs() {
			panic(fmt.Errorf("given URL %q must be absolute but it is not", baseURLs[i]))
		}
	}

	pool := hostpool.NewEpsilonGreedy(baseURLs, decayDuration, valueCalculator)
	clients := make(map[string]Client)
	mu := &sync.Mutex{}

	return func(c Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			r := pool.Get()
			h := r.Host()

			mu.Lock()
			pooledClient, ok := clients[h]
			if !ok {
				pooledClient = Decorate(c, Scoped(h))
				clients[h] = pooledClient
			}
			mu.Unlock()

			resp, err := pooledClient.Do(req)
			r.Mark(err)

			return resp, err
		})
	}
}
