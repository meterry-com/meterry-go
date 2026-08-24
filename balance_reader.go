package sdk

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/meterry-com/meterry-go/pkg/types"
)

const (
	defaultBalanceCacheTTL         = 3 * time.Second
	defaultNegativeBalanceCacheTTL = time.Second
)

// BalanceReaderOptions controls the optional in-process cache used by a
// BalanceReader. The direct Manager read methods remain uncached.
type BalanceReaderOptions struct {
	CacheTTL         time.Duration
	NegativeCacheTTL time.Duration
	RequestTimeout   time.Duration
}

type balanceCacheEntry struct {
	snapshot  types.VirtualWalletAmountSnapshot
	expiresAt time.Time
}

type balanceReadCall struct {
	done     chan struct{}
	snapshot *types.VirtualWalletAmountSnapshot
	err      error
}

// BalanceReader provides an opt-in cached and concurrency-coalesced view of
// realtime wallet amounts for one project. It does not apply billing policy,
// reserve funds, or cache errors.
type BalanceReader struct {
	client         *Client
	projectID      string
	cacheTTL       time.Duration
	negativeTTL    time.Duration
	requestTimeout time.Duration
	mu             sync.Mutex
	cache          map[string]balanceCacheEntry
	inFlight       map[string]*balanceReadCall
}

// NewBalanceReader creates an opt-in balance reader for projectID.
func (c *Client) NewBalanceReader(projectID string, options BalanceReaderOptions) (*BalanceReader, error) {
	if c == nil {
		return nil, errors.New("billing sdk client is nil")
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, errors.New("billing balance reader project id is required")
	}
	if options.CacheTTL <= 0 {
		options.CacheTTL = defaultBalanceCacheTTL
	}
	if options.NegativeCacheTTL <= 0 {
		options.NegativeCacheTTL = defaultNegativeBalanceCacheTTL
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = c.httpClient.Timeout
	}
	return &BalanceReader{
		client:         c,
		projectID:      projectID,
		cacheTTL:       options.CacheTTL,
		negativeTTL:    options.NegativeCacheTTL,
		requestTimeout: options.RequestTimeout,
		cache:          make(map[string]balanceCacheEntry),
		inFlight:       make(map[string]*balanceReadCall),
	}, nil
}

// ReadVirtualWalletAmount returns the realtime wallet amount, reusing a
// recent result and coalescing concurrent reads for the same request key.
func (r *BalanceReader) ReadVirtualWalletAmount(ctx context.Context, req types.ReadVirtualWalletRequest) (*types.VirtualWalletAmountSnapshot, error) {
	if r == nil || r.client == nil {
		return nil, errors.New("billing balance reader is nil")
	}
	key := balanceReadKey(req)
	now := time.Now()
	r.mu.Lock()
	if entry, ok := r.cache[key]; ok && now.Before(entry.expiresAt) {
		r.mu.Unlock()
		return cloneBalanceSnapshot(entry.snapshot), nil
	}
	if call, ok := r.inFlight[key]; ok {
		r.mu.Unlock()
		select {
		case <-call.done:
			if call.snapshot == nil {
				return nil, call.err
			}
			return cloneBalanceSnapshot(*call.snapshot), call.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	call := &balanceReadCall{done: make(chan struct{})}
	r.inFlight[key] = call
	r.mu.Unlock()

	readCtx := context.WithoutCancel(ctx)
	cancel := func() {}
	if r.requestTimeout > 0 {
		readCtx, cancel = context.WithTimeout(readCtx, r.requestTimeout)
	}
	call.snapshot, call.err = r.client.Manager.ReadVirtualWalletAmountForProject(readCtx, r.projectID, req)
	cancel()
	r.mu.Lock()
	if call.err == nil && call.snapshot != nil {
		ttl := r.cacheTTL
		if call.snapshot.Balance.Sign() <= 0 {
			ttl = r.negativeTTL
		}
		r.cache[key] = balanceCacheEntry{snapshot: *call.snapshot, expiresAt: time.Now().Add(ttl)}
	}
	delete(r.inFlight, key)
	close(call.done)
	r.mu.Unlock()
	if call.snapshot == nil {
		return nil, call.err
	}
	return cloneBalanceSnapshot(*call.snapshot), call.err
}

// Invalidate removes a cached amount for the given request key. It is useful
// when an application receives a wallet or balance webhook.
func (r *BalanceReader) Invalidate(req types.ReadVirtualWalletRequest) {
	if r == nil {
		return
	}
	r.mu.Lock()
	delete(r.cache, balanceReadKey(req))
	r.mu.Unlock()
}

func balanceReadKey(req types.ReadVirtualWalletRequest) string {
	return strings.Join([]string{
		strings.TrimSpace(req.AccountID),
		strings.TrimSpace(req.SubjectType),
		strings.TrimSpace(req.SubjectID),
		strings.TrimSpace(req.Currency),
	}, "\x00")
}

func cloneBalanceSnapshot(snapshot types.VirtualWalletAmountSnapshot) *types.VirtualWalletAmountSnapshot {
	return &snapshot
}
