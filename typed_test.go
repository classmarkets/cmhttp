package cmhttp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(contentType, accept string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != contentType {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Expected Content-Type %q, got %q", contentType, r.Header.Get("Content-Type"))
			return
		}

		if r.Header.Get("Accept") != accept {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Expected Accept %q, got %q", accept, r.Header.Get("Accept"))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}))
}

func TestRequestHeaders(t *testing.T) {
	server := testServer("application/json", "application/json")
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

	if r.StatusCode != http.StatusAccepted {
		s, _ := ioutil.ReadAll(r.Body)
		t.Errorf("Unexpected response status %d: %s", r.StatusCode, s)
	}
}

func TestRequestHeadersOverride(t *testing.T) {
	server := testServer("image/jpeg", "image/png")
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

	if r.StatusCode != http.StatusAccepted {
		s, _ := ioutil.ReadAll(r.Body)
		t.Errorf("Unexpected response status %d: %s", r.StatusCode, s)
	}
}
