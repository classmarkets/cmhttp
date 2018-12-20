package cmhttp

import (
	"io"
	"io/ioutil"
)

// DrainClose reads up to 256 kB from body and then calls Close. If draining
// the body returns a non-nil error other than io.EOF, that error is returned.
// Otherwise the error from Close() is returned. Close() is called even if draining fails.
//
// If body is nil, DrainClose returns a nil error immediately.
func DrainClose(body io.ReadCloser) error {
	if body == nil {
		return nil
	}

	// 256k is the value of http.maxPostHandlerReadBytes, "approximately what a
	// typical machine's TCP buffer size is".
	_, err := io.CopyN(ioutil.Discard, body, 256<<10)
	if err != nil && err != io.EOF {
		body.Close()
		return err
	}

	return body.Close()
}
