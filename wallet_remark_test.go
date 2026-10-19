package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/meterry-com/meterry-go/pkg/types"
	"github.com/shopspring/decimal"
)

func TestWalletCreditRemarkRoundTrip(t *testing.T) {
	for _, method := range []string{"CreditWallet", "TopUpWallet"} {
		for _, remark := range []string{"", "补录客户银行转账"} {
			t.Run(method+"/"+remark, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodPost || r.URL.Path != "/v1/projects/project/wallets/credit" {
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					}
					var body map[string]json.RawMessage
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					var receivedRemark, sourceID string
					if value, ok := body["remark"]; ok {
						if err := json.Unmarshal(value, &receivedRemark); err != nil {
							t.Error(err)
						}
					} else if remark != "" {
						t.Error("remark missing from request")
					}
					if remark == "" {
						if _, ok := body["remark"]; ok {
							t.Error("empty remark should be omitted")
						}
					}
					if err := json.Unmarshal(body["source_id"], &sourceID); err != nil {
						t.Error(err)
					}
					if receivedRemark != remark || sourceID != "payment_001" {
						t.Errorf("remark=%q source_id=%q", receivedRemark, sourceID)
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(types.CreditWalletResponse{LedgerEntry: types.WalletLedgerEntry{ID: "ledger", Remark: receivedRemark, SourceID: sourceID}})
				}))
				defer server.Close()
				client := MustNewClient(Config{BaseURL: server.URL, ProjectID: "project", APIKey: "key"})
				req := types.CreditWalletRequest{AccountID: "account", Currency: "USD", Amount: decimal.NewFromInt(100), SourceID: "payment_001", Remark: remark}
				credit := client.Manager.CreditWallet
				if method == "TopUpWallet" {
					credit = client.Manager.TopUpWallet
				}
				response, err := credit(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				if response.LedgerEntry.Remark != remark || response.LedgerEntry.SourceID != "payment_001" {
					t.Fatalf("response = %#v", response)
				}
			})
		}
	}
}

func TestListWalletLedgerEntriesReadsRemark(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/projects/project/wallet-ledger-entries" || r.URL.Query().Get("account_id") != "account" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ledger_entries":[{"id":"new","source_id":"payment_001","remark":"补录客户银行转账"},{"id":"legacy","source_id":"legacy note"}],"total_count":2,"limit":50,"offset":0}`))
	}))
	defer server.Close()
	client := MustNewClient(Config{BaseURL: server.URL, ProjectID: "project", APIKey: "key"})
	response, err := client.Manager.ListWalletLedgerEntries(context.Background(), types.ListWalletLedgerEntriesRequest{AccountID: "account"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.LedgerEntries) != 2 {
		t.Fatalf("response = %#v", response)
	}
	if response.LedgerEntries[0].Remark != "补录客户银行转账" || response.LedgerEntries[0].SourceID != "payment_001" {
		t.Fatalf("new entry = %#v", response.LedgerEntries[0])
	}
	if response.LedgerEntries[1].Remark != "" || response.LedgerEntries[1].SourceID != "legacy note" {
		t.Fatalf("legacy entry = %#v", response.LedgerEntries[1])
	}
}
