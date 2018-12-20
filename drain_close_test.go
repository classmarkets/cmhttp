package cmhttp

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDrainClose_NilBody(t *testing.T) {
	if err := DrainClose(nil); err != nil {
		t.Fatal("DrainClose(nil) returned non-nil error:", err)
	}
}

func TestDrainClose_AtEOF(t *testing.T) {
	buf := &bytes.Buffer{}
	body := ioutil.NopCloser(buf)

	if err := DrainClose(body); err != nil {
		t.Fatal("DrainClose() returned non-nil error for already drained body:", err)
	}
}

type testReadCloser struct {
	readErr  error
	closeErr error
	closed   bool
}

func (r *testReadCloser) Read([]byte) (int, error) {
	return 0, r.readErr
}
func (r *testReadCloser) Close() error {
	r.closed = true
	return r.closeErr
}

func TestDrainClose_AlwaysCloses(t *testing.T) {
	body := &testReadCloser{
		readErr:  errors.New("read fail!"),
		closeErr: errors.New("close fail!"),
	}

	err := DrainClose(body)
	if err == nil || err.Error() != "read fail!" {
		t.Error("DrainClose() didn't return read error:", err)
	}

	if !body.closed {
		t.Error("DrainClose() didn't Close() after read error")
	}
}

func TestDrainClose_StillRequired(t *testing.T) {
	nDialsControl := testDrainClose(t, false, 10)
	nDials := testDrainClose(t, true, 10)

	if nDialsControl <= nDials {
		t.Error("DrainClose(resp.Body) doesn't result in fewer dials than resp.Body.Close(). It may not be necessary anymore")
	}
}

func testDrainClose(t *testing.T, drain bool, nReq int) int32 {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}\n\n"))
	}))

	var n int32
	c := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				atomic.AddInt32(&n, 1)
				return net.Dial(network, addr)
			},
		},
	}

	for i := 0; i < nReq; i++ {
		req, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if drain {
			DrainClose(res.Body)
		} else {
			res.Body.Close()
		}

		// http.Transport does manages connections in goroutines, so we give
		// them a chance to recycle the connection we just used.
		time.Sleep(10 * time.Millisecond)
	}

	return n
}
