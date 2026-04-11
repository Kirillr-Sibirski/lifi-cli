package earn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientRetriesOnServerFailure(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		http:    &http.Client{Timeout: time.Second},
	}
	chains, err := client.GetChains(context.Background())
	if err != nil {
		t.Fatalf("GetChains returned error: %v", err)
	}
	if len(chains) != 0 {
		t.Fatalf("expected no chains, got %d", len(chains))
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}
