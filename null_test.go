package cmhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("A Null client should never send a request")
	}))
	defer server.Close()

	client := Null()(http.DefaultClient)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	r, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	if r.StatusCode != http.StatusNoContent {
		t.Errorf("Unexpected response status %d", r.StatusCode)
	}
}
