package cmhttp

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var metrics struct {
	reqCnt *prometheus.CounterVec
	reqDur prometheus.Summary
	reqSz  prometheus.Summary
	resSz  prometheus.Summary
}

// Instrumented registers and collects four prometheus metrics on the decorated HTTP client.
// The following four metric collectors are registered (if not already done):
//     - http_requests_total (CounterVec)
//     - http_request_duration_microseconds (Summary),
//     - http_request_size_bytes (Summary)
//     - http_response_size_bytes (Summary)
// Each has a constant label named "name" with the provided name as value.
// http_requests_total is a metric vector partitioned by HTTP method
// (label name "method") and HTTP status code (label name "code").
//
// This code closely resembles the HTTP server side prometheus.InstrumentHandler function.
func Instrumented(name string) Decorator {
	return InstrumentedWithOpts(
		prometheus.SummaryOpts{
			Subsystem:   "http_client",
			ConstLabels: prometheus.Labels{"name": name},
		},
	)
}

// InstrumentedWithOpts is like Instrumented but gives you direct control over the used
// prometheus.SummaryOpts.
func InstrumentedWithOpts(opts prometheus.SummaryOpts) Decorator {
	if metrics.reqCnt == nil {
		// make sure we register metrics only once
		metrics.reqCnt = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   opts.Namespace,
				Subsystem:   opts.Subsystem,
				Name:        "requests_total",
				Help:        "Total number of HTTP requests made.",
				ConstLabels: opts.ConstLabels,
			},
			[]string{"method", "code"},
		)

		opts.Name = "request_duration_microseconds"
		opts.Help = "The HTTP request latencies in microseconds."
		metrics.reqDur = prometheus.NewSummary(opts)

		opts.Name = "request_size_bytes"
		opts.Help = "The HTTP request sizes in bytes."
		metrics.reqSz = prometheus.NewSummary(opts)

		opts.Name = "response_size_bytes"
		opts.Help = "The HTTP response sizes in bytes."
		metrics.resSz = prometheus.NewSummary(opts)

		prometheus.MustRegister(metrics.reqCnt)
		prometheus.MustRegister(metrics.reqDur)
		prometheus.MustRegister(metrics.reqSz)
		prometheus.MustRegister(metrics.resSz)
	}

	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			begin := time.Now()
			resp, err := c.Do(r)
			if err != nil {
				return resp, err
			}

			elapsed := float64(time.Since(begin)) / float64(time.Microsecond)
			go func() {
				requestSize := computeApproximateRequestSize(r)
				method := sanitizeMethod(r.Method)
				code := sanitizeCode(resp.StatusCode)
				respLengthHeader := resp.Header.Get("Content-Length")
				respLength, _ := strconv.Atoi(respLengthHeader) // we can't do anything in case this is not an integer so we ignore this case

				metrics.reqCnt.WithLabelValues(method, code).Inc()
				metrics.reqDur.Observe(elapsed)
				metrics.reqSz.Observe(float64(respLength))
				metrics.resSz.Observe(float64(requestSize))
			}()

			return resp, err
		})
	}
}

// computeApproximateRequestSize has been mostly copied from the prometheus.InstrumentHandler logic.
func computeApproximateRequestSize(r *http.Request) int {
	var s int
	if r.URL != nil {
		s = len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}

	return s
}

// sanitizeMethod normalizes HTTP methods to lowercase.
// This was copied from prometheus.InstrumentHandler logic.
func sanitizeMethod(m string) string {
	switch m {
	case "GET", "get":
		return "get"
	case "PUT", "put":
		return "put"
	case "HEAD", "head":
		return "head"
	case "POST", "post":
		return "post"
	case "DELETE", "delete":
		return "delete"
	case "CONNECT", "connect":
		return "connect"
	case "OPTIONS", "options":
		return "options"
	case "NOTIFY", "notify":
		return "notify"
	default:
		return strings.ToLower(m)
	}
}

// sanitizeCode maps HTTP response status codes to a string representation.
// This was copied from prometheus.InstrumentHandler logic.
func sanitizeCode(s int) string {
	switch s {
	case 100:
		return "100"
	case 101:
		return "101"

	case 200:
		return "200"
	case 201:
		return "201"
	case 202:
		return "202"
	case 203:
		return "203"
	case 204:
		return "204"
	case 205:
		return "205"
	case 206:
		return "206"

	case 300:
		return "300"
	case 301:
		return "301"
	case 302:
		return "302"
	case 304:
		return "304"
	case 305:
		return "305"
	case 307:
		return "307"

	case 400:
		return "400"
	case 401:
		return "401"
	case 402:
		return "402"
	case 403:
		return "403"
	case 404:
		return "404"
	case 405:
		return "405"
	case 406:
		return "406"
	case 407:
		return "407"
	case 408:
		return "408"
	case 409:
		return "409"
	case 410:
		return "410"
	case 411:
		return "411"
	case 412:
		return "412"
	case 413:
		return "413"
	case 414:
		return "414"
	case 415:
		return "415"
	case 416:
		return "416"
	case 417:
		return "417"
	case 418:
		return "418"

	case 500:
		return "500"
	case 501:
		return "501"
	case 502:
		return "502"
	case 503:
		return "503"
	case 504:
		return "504"
	case 505:
		return "505"

	case 428:
		return "428"
	case 429:
		return "429"
	case 431:
		return "431"
	case 511:
		return "511"

	default:
		return strconv.Itoa(s)
	}
}

// InstrumentedRequestDurations instruments the client by tracking the request
// durations as histogram vector partitioned by HTTP method (label name "method")
// and response code (label name "code").
func InstrumentedRequestDurations(opts prometheus.HistogramOpts) Decorator {
	opts.Name = "request_duration_microseconds"
	opts.Help = "The HTTP request duration in microseconds."

	durations := prometheus.NewHistogramVec(opts, []string{"method", "code"})
	prometheus.MustRegister(durations)

	return func(c Client) Client {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			begin := time.Now()
			resp, err := c.Do(r)
			if err != nil {
				return resp, err
			}

			elapsed := float64(time.Since(begin)) / float64(time.Microsecond)

			method := sanitizeMethod(r.Method)
			code := sanitizeCode(resp.StatusCode)
			durations.WithLabelValues(method, code).Observe(elapsed)

			return resp, err
		})
	}
}
