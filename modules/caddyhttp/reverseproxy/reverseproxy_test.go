package reverseproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

type blockReader struct {
	ctx context.Context
}

func (b *blockReader) Read(p []byte) (n int, err error) {
	select {
	case <-b.ctx.Done():
		return 0, b.ctx.Err()
	}
}

func (b *blockReader) Close() error {
	return nil
}

func TestClientDisconnectLeak(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	rt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       &blockReader{ctx: req.Context()},
			}, nil
		},
	}

	proxy := ReverseProxy{}
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel() // Simulate client disconnect
	}()

	err := proxy.ServeHTTP(w, req, rt)
	if err != nil && err != context.Canceled {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
