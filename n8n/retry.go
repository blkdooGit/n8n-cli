package n8n

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig is the default retry configuration
var DefaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	InitialWait: 1 * time.Second,
	MaxWait:     30 * time.Second,
	Multiplier:  2.0,
}

// RetryTransport is an http.RoundTripper that retries on transient errors
type RetryTransport struct {
	Base   http.RoundTripper
	Config RetryConfig
	Logger *zap.SugaredLogger
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}

	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= t.Config.MaxRetries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(float64(t.Config.InitialWait) * math.Pow(t.Config.Multiplier, float64(attempt-1)))
			if wait > t.Config.MaxWait {
				wait = t.Config.MaxWait
			}
			if t.Logger != nil {
				t.Logger.Debugf("Retry attempt %d/%d after %v", attempt, t.Config.MaxRetries, wait)
			}
			time.Sleep(wait)
		}

		reqClone := req.Clone(req.Context())
		if bodyBytes != nil {
			reqClone.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			reqClone.ContentLength = int64(len(bodyBytes))
		}

		var err error
		resp, err = t.Base.RoundTrip(reqClone)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
			resp.Body.Close()
			lastErr = nil
			continue
		}

		return resp, nil
	}

	return resp, lastErr
}
