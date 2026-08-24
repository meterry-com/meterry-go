package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestBalanceReaderCachesAndCoalescesConcurrentReads(t *testing.T) {
	var requests atomic.Int32
	started := make(chan struct{})
	continueRequest := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			close(started)
			<-continueRequest
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"balance": "1"})
	}))
	defer server.Close()

	client := MustNewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	reader, err := client.NewBalanceReader("proj", BalanceReaderOptions{
		CacheTTL:       time.Minute,
		RequestTimeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	const callers = 20
	results := make(chan error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			snapshot, readErr := reader.ReadVirtualWalletAmount(context.Background(), types.ReadVirtualWalletRequest{
				SubjectType: "api_key",
				SubjectID:   "client-a",
				Currency:    "USD",
			})
			if readErr != nil || snapshot == nil || snapshot.Balance.String() != "1" {
				results <- readErr
				return
			}
			results <- nil
		}()
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for balance request")
	}
	close(continueRequest)
	wg.Wait()
	close(results)
	for readErr := range results {
		if readErr != nil {
			t.Fatal(readErr)
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("requests=%d, want 1", got)
	}
	if _, err := reader.ReadVirtualWalletAmount(context.Background(), types.ReadVirtualWalletRequest{
		SubjectType: "api_key", SubjectID: "client-a", Currency: "USD",
	}); err != nil {
		t.Fatal(err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("cached requests=%d, want 1", got)
	}
}

func TestBalanceReaderInvalidate(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":"1"}`))
	}))
	defer server.Close()

	client := MustNewClient(Config{BaseURL: server.URL, APIKey: "secret"})
	reader, err := client.NewBalanceReader("proj", BalanceReaderOptions{CacheTTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	req := types.ReadVirtualWalletRequest{SubjectType: "api_key", SubjectID: "client-a", Currency: "USD"}
	if _, err := reader.ReadVirtualWalletAmount(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	reader.Invalidate(req)
	if _, err := reader.ReadVirtualWalletAmount(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("requests after invalidate=%d, want 2", got)
	}
}
