package cmhttp

import (
	"net/http"
	"time"
	"sync"

	"github.com/bitly/go-hostpool"
)

// ClientPool creates a pool of HTTP clients that use an ε-greedy strategy to distribute
// HTTP requests among multiple hosts. The pool transparently distributes the requests
// among the hosts taking the individual request durations and failures into account.
// A more detailed discussion of the underlying algorithm can be found at
// https://godoc.org/github.com/bitly/go-hostpool#NewEpsilonGreedy
//
// See also https://en.wikipedia.org/wiki/Epsilon-greedy_strategy
func ClientPool(hosts []string, decayDuration time.Duration, valueCalculator hostpool.EpsilonValueCalculator) Decorator {
	pool := hostpool.NewEpsilonGreedy(hosts, decayDuration, valueCalculator)
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
