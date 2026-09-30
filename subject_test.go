package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/meterry-com/meterry-go/pkg/types"
)

func TestGrantSubjectContract(t *testing.T) {
	for _, restricted := range []bool{false, true} {
		name := "defaults"
		if restricted {
			name = "restricted metrics and explicit lifetime"
		}
		t.Run(name, func(t *testing.T) {
			req := types.GrantSubjectRequest{
				SubjectType: "user", SubjectID: "user_123", Rules: []string{"analysis"},
			}
			wantBody := map[string]any{
				"subject_type": "user", "subject_id": "user_123", "rules": []any{"analysis"},
			}
			want := types.GrantSubjectResponse{
				TokenType: "Bearer", AccessToken: "subject.jwt.token",
				ExpiresIn: 900, ExpiresAt: 1790756100, Rules: []string{"analysis"},
			}
			want.Subject.Type, want.Subject.ID = "user", "user_123"
			if restricted {
				req.Metrics = []string{"input_tokens", "output_tokens"}
				req.ExpiresIn = 600
				wantBody["metrics"] = []any{"input_tokens", "output_tokens"}
				wantBody["expires_in"] = float64(600)
				want.Metrics = req.Metrics
				want.ExpiresIn = 600
				want.ExpiresAt = 1790755800
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/gateway/v1/subject/grants" || r.URL.RawQuery != "" {
					t.Errorf("request = %s %s", r.Method, r.URL.RequestURI())
				}
				if r.Header.Get("Authorization") != "Bearer tenant_key" {
					t.Errorf("authorization = %q", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("X-Tenant-ID") != "" {
					t.Errorf("headers = %v", r.Header)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(body, wantBody) {
					t.Errorf("body = %#v; want %#v", body, wantBody)
				}
				w.Header().Set("Content-Type", "application/json")
				// Use the server's wire schema independently of the SDK DTO.
				payload := map[string]any{
					"token_type": "Bearer", "access_token": "subject.jwt.token",
					"expires_in": want.ExpiresIn, "expires_at": want.ExpiresAt,
					"subject": map[string]string{"type": "user", "id": "user_123"},
					"rules":   []string{"analysis"},
				}
				if restricted {
					payload["metrics"] = []string{"input_tokens", "output_tokens"}
				}
				if err := json.NewEncoder(w).Encode(payload); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL + "/gateway/", APIKey: "tenant_key"})
			got, err := client.Manager.GrantSubject(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, &want) {
				t.Fatalf("response = %#v; want %#v", got, &want)
			}
		})
	}
}

func TestGrantSubjectReturnsAPIError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		message string
	}{
		{"invalid grant", http.StatusBadRequest, "unknown subject rule"},
		{"invalid credentials", http.StatusUnauthorized, "invalid api key"},
		{"missing scope", http.StatusForbidden, "forbidden scope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": tc.message})
			}))
			defer server.Close()
			client := MustNewClient(Config{ProjectID: "proj_001", BaseURL: server.URL, APIKey: "tenant_key"})
			_, err := client.Manager.GrantSubject(context.Background(), types.GrantSubjectRequest{
				SubjectType: "user", SubjectID: "user_123", Rules: []string{"analysis"},
			})
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status || apiErr.Message != tc.message {
				t.Fatalf("error = %v; want APIError(%d, %q)", err, tc.status, tc.message)
			}
		})
	}
}
