package cmhttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Expected Content-Type %q, got %q", "application/json", r.Header.Get("Content-Type"))
			return
		}

		if r.Header.Get("Accept") != "application/json" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Expected Accept %q, got %q", "application/json", r.Header.Get("Accept"))
			return
		}

		w.WriteHeader(615)
	}))
	defer server.Close()

	client := Typed("application/json")(http.DefaultClient)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}

	r, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != 615 {
		s, _ := ioutil.ReadAll(r.Body)
		t.Errorf("Unexpected response status %d: %s", r.StatusCode, s)
	}
}

func TestRequestHeadersOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "image/jpeg" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Expected Content-Type %q, got %q", "image/jpeg", r.Header.Get("Content-Type"))
			return
		}

		if r.Header.Get("Accept") != "image/png" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Expected Accept %q, got %q", "image/png", r.Header.Get("Accept"))
			return
		}

		w.WriteHeader(615)
	}))
	defer server.Close()

	client := JSON()(http.DefaultClient)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "image/jpeg")
	req.Header.Set("Accept", "image/png")

	r, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != 615 {
		s, _ := ioutil.ReadAll(r.Body)
		t.Errorf("Unexpected response status %d: %s", r.StatusCode, s)
	}
}
