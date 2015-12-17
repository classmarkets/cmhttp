package cmhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRelativeURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/test" {
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := Scoped(server.URL)(http.DefaultClient)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Error(err)
	}
	r, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != http.StatusAccepted {
		t.Errorf("Did not reach test server")
	}
}

func TestAbsoluteURL(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/test1" {
			w.WriteHeader(614)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/test2" {
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server2.Close()

	client := Scoped(server1.URL)(http.DefaultClient)

	req, err := http.NewRequest("GET", server2.URL+"/test2", nil)
	if err != nil {
		t.Error(err)
	}
	r, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != http.StatusAccepted {
		t.Errorf("Did not reach correct test server")
	}
}
