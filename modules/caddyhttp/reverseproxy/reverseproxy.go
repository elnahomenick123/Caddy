package reverseproxy

import (
	"context"
	"io"
	"net/http"
)

// ReverseProxy handles proxying requests to upstreams.
type ReverseProxy struct {
	FlushInterval int
}

func (h ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, rt http.RoundTripper) error {
	outreq := r.Clone(r.Context())

	res, err := rt.RoundTrip(outreq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Monitor client disconnect to close upstream response body immediately
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		<-ctx.Done()
		res.Body.Close()
	}()

	// Copy response body
	_, err = io.Copy(w, res.Body)
	return err
}
