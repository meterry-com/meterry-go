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

func TestVoucherGrantRemarkRoundTrip(t *testing.T) {
	for _, remark := range []string{"", "夏季活动赠送额度"} {
		t.Run(remark, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path != "/v1/projects/project/voucher-grants" {
					t.Errorf("path = %s", r.URL.Path)
				}
				grant := types.VoucherGrant{ID: "voucher", SourceID: "promotion_001"}
				if r.Method == http.MethodGet {
					_ = json.NewEncoder(w).Encode(types.ListVoucherGrantsResponse{VoucherGrants: []types.VoucherGrant{grant, {ID: "legacy"}}})
					return
				}
				if r.Method != http.MethodPost {
					t.Errorf("method = %s", r.Method)
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				var req types.CreateVoucherGrantRequest
				encoded, _ := json.Marshal(body)
				if err := json.Unmarshal(encoded, &req); err != nil {
					t.Error(err)
				}
				if req.Remark != remark || req.SourceID != "promotion_001" {
					t.Errorf("request = %#v", req)
				}
				_, present := body["remark"]
				if present != (remark != "") {
					t.Errorf("remark presence = %t", present)
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(types.CreateVoucherGrantResponse{VoucherGrant: grant, LedgerEntry: types.WalletLedgerEntry{Remark: remark}})
			}))
			defer server.Close()
			client := MustNewClient(Config{BaseURL: server.URL, ProjectID: "project", APIKey: "key"})
			response, err := client.Manager.CreateVoucherGrant(context.Background(), types.CreateVoucherGrantRequest{AccountID: "account", Currency: "USD", Amount: decimal.NewFromInt(10), SourceID: "promotion_001", Remark: remark})
			if err != nil {
				t.Fatal(err)
			}
			if response.LedgerEntry.Remark != remark {
				t.Fatalf("response = %#v", response)
			}
			list, err := client.Manager.ListVoucherGrants(context.Background(), ListVoucherGrantsRequest{AccountID: "account"})
			if err != nil {
				t.Fatal(err)
			}
			if len(list.VoucherGrants) != 2 || list.VoucherGrants[0].SourceID != "promotion_001" || list.VoucherGrants[1].ID != "legacy" {
				t.Fatalf("list = %#v", list)
			}
		})
	}
}
