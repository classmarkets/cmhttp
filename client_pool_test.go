package cmhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"math/rand"

	"github.com/bitly/go-hostpool"
	"github.com/classmarkets/cmhttp"
)

func init() {
	rand.Seed(42)
}

func TestClientPool(t *testing.T) {
	decayDuration := 1 * time.Second // very short interval just for the test
	valueCalculator := new(hostpool.LinearEpsilonValueCalculator)

	handlers := map[string]http.Handler{
		"fast": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
		}),
		"mediocre": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(25 * time.Millisecond)
		}),
		"slow": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(50 * time.Millisecond)
		}),
	}

	stats := make(map[string]int)
	var hosts []string
	for name := range handlers {
		name := name // redefine in this scope so it doesn't get overwritten in the next loop iteration
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers[name].ServeHTTP(w, r)
			stats[name] = stats[name] + 1
		}))
		hosts = append(hosts, s.URL)
	}

	c := cmhttp.Decorate(http.DefaultClient,
		cmhttp.JSON(),
		cmhttp.ClientPool(hosts, decayDuration, valueCalculator),
	)

	for i := 0; i < 100; i++ {
		req, _ := http.NewRequest("GET", "/", nil)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	for name := range handlers {
		t.Logf("%s handler received %d requests", name, stats[name])
		if stats[name] == 0 {
			t.Errorf("Each handler should have gotten at least one request but %s has none (0)", name)
		}
	}

	if stats["slow"] > stats["mediocre"] {
		t.Errorf("Slow handler received more requests than the mediocre handler")
	}

	if stats["mediocre"] > stats["fast"] {
		t.Errorf("Mediocre handler received more requests than the fast handler")
	}
}
