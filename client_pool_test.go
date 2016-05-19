package cmhttp_test

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bitly/go-hostpool"
	"github.com/classmarkets/cmhttp"
)

func init() {
	rand.Seed(42)
}

func TestStaticClientPool(t *testing.T) {
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

	stats := runStaticClientPoolTest(t, 100, handlers)

	for name := range handlers {
		t.Logf("%s handler received %d requests", name, stats[name])
		if stats[name] == 0 {
			t.Errorf("Each handler should have gotten at least one request but %s has none (0)", name)
		}
	}

	if stats["slow"] > stats["mediocre"] {
		t.Error("Slow handler received more requests than the mediocre handler")
	}

	if stats["mediocre"] > stats["fast"] {
		t.Error("Mediocre handler received more requests than the fast handler")
	}
}

func TestStaticClientPool_WhenSomeServersFail(t *testing.T) {
	handlers := map[string]http.Handler{
		"ok": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
		}),
		"fail1": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
		"fail2": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	}

	stats := runStaticClientPoolTest(t, 100, handlers)
	for name := range handlers {
		t.Logf("%s handler received %d requests", name, stats[name])
		if stats[name] == 0 {
			t.Errorf("Each handler should have gotten at least one request but %s has none (0)", name)
		}
	}

	if stats["fail1"] > stats["ok"] {
		t.Error("Failure handler 1 received more requests than the ok handler")
	}

	if stats["fail2"] > stats["ok"] {
		t.Error("Failure handler 2 received more requests than the ok handler")
	}
}

func TestStaticClientPool_WhenAllServersFail(t *testing.T) {
	handlers := map[string]http.Handler{
		"s1": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
		"s2": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
		"s3": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	}

	n := 100
	stats := runStaticClientPoolTest(t, n, handlers)
	var total int
	for _, s := range stats {
		total += s
	}

	if total != n {
		t.Errorf("Total number of made requests == %d, want %d", total, n)
	}
}

func runStaticClientPoolTest(t *testing.T, n int, handlers map[string]http.Handler) (stats map[string]int) {
	decayDuration := 1 * time.Second // very short interval just for the test
	valueCalculator := new(hostpool.LinearEpsilonValueCalculator)

	stats = make(map[string]int)
	var urls []string
	for name := range handlers {
		name := name // redefine in this scope so it doesn't get overwritten in the next loop iteration
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers[name].ServeHTTP(w, r)
			stats[name] = stats[name] + 1
		}))
		urls = append(urls, s.URL)
	}

	c := cmhttp.Decorate(http.DefaultClient,
		cmhttp.JSON(),
		cmhttp.StaticClientPool(urls, decayDuration, valueCalculator),
	)

	for i := 0; i < n; i++ {
		req, _ := http.NewRequest("GET", "/", nil)
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	return stats
}

func TestStaticClientPool_PanicIfInvalidURLsAreGiven(t *testing.T) {
	d := 1 * time.Second // doesn't actually matter in this test at all
	vc := new(hostpool.LinearEpsilonValueCalculator)

	cases := []struct {
		url         string
		shouldPanic bool
	}{
		{"", true},
		{"foobar", true},
		{"foo.bar", true},
		{"foo.bar:8080", false},
		{"http://foobar", false},
		{"https://foobar.com", false},
		{"http://foobar.com:8080", false},
	}

	for _, c := range cases {
		paniced := checkPanic(t, func() {
			cmhttp.StaticClientPool([]string{c.url}, d, vc)
		})

		if paniced != c.shouldPanic {
			if c.shouldPanic {
				t.Errorf("URL %q : should have paniced", c.url)
			} else {
				t.Errorf("URL %q : should not have paniced", c.url)
			}
		}
	}
}

func checkPanic(t *testing.T, fun func()) (panics bool) {
	defer func() {
		r := recover()
		if r != nil {
			panics = true
		}
	}()

	fun()
	return
}
