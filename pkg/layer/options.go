package layer

import (
	"context"
	"errors"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"io"
	"net"
	"net/http"
	"syscall"
	"time"
)

var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	MaxIdleConnsPerHost:   50,
}

var defaultRetryStatusCodes = []int{
	http.StatusRequestTimeout,
	http.StatusInternalServerError,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	499,
}

func ShouldRetry(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) || errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, net.ErrClosed) {
		return true
	}

	return false
}

func GetTransport() http.RoundTripper {
	return transport.NewRetry(
		DefaultTransport,
		transport.WithRetryPredicate(ShouldRetry),
		transport.WithRetryStatusCodes(defaultRetryStatusCodes...),
	)
}
