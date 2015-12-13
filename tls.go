package cmhttp

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

// ConfigureTLS reads the certificate bundle at rootCertsFilePath and
// configures the given transport to verify server certificate against these
// certificates.
func ConfigureTLS(t *http.Transport, rootCertsFilePath string) error {
	certPool := x509.NewCertPool()
	buf, err := ioutil.ReadFile(rootCertsFilePath)
	if err != nil {
		return err
	}

	certPool.AppendCertsFromPEM(buf)
	t.TLSClientConfig = &tls.Config{RootCAs: certPool}

	return nil
}

// MustConfigureTLS is the same as ConfigureTLS, but panics if there is an
// error.
func MustConfigureTLS(t *http.Transport, rootCertsFilePath string) {
	if err := ConfigureTLS(t, rootCertsFilePath); err != nil {
		panic(err)
	}
}
