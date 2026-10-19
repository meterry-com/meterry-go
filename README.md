# meterry-go

Go SDK for Meterry.

## Install

```bash
go get github.com/meterry-com/meterry-go
```

## Setup

Create a client for your project using your Meterry API key:

```go
package main

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/meterry-com/meterry-go"
)

func main() {
	client, err := sdk.NewClient(sdk.Config{
		ProjectID: os.Getenv("METERRY_PROJECT_ID"),
		BaseURL:   os.Getenv("METERRY_BASE_URL"),
		APIKey:    os.Getenv("METERRY_API_KEY"),
	})
	if err != nil {
		panic(err)
	}

	_ = context.Background()
	fmt.Println(client)
}
```

`Config` supports:

- `BaseURL`: required, for example `https://api.meterry.com`
- `ProjectID`: required project ID
- `APIKey`: required bearer token identifying the tenant and authorized for the project
- `HTTPClient`: optional custom HTTP client
- `UserAgent`: optional custom user agent

No tenant ID is needed at initialization. Tenant identity is authenticated by the API key; the SDK does not send a tenant header. Project-scoped operations use the configured project automatically. Subject token grants derive the project from the API key, which must be bound to that project. Create a separate client for each tenant/project pair.

If you prefer a panic-on-error constructor, use `sdk.MustNewClient(...)`.

## Usage

The client exposes three main entry points:

- `client.Manager`
- `client.Query`
- `client.Ingest`

`client.Manager` includes project-scoped helpers for wallets, usage-control rules, and the subject realtime limits snapshot at `/v1/projects/:project_id/subject/realtime/limits`.
`ReadRealtimeWalletLimits` and `ReadVirtualWalletLimits` are deprecated wrappers that preserve the legacy wallet-limits snapshot shape while sourcing from `ReadSubjectRealtimeLimits`.
`ReadVirtualWalletAmount` remains the realtime balance helper.

For example, create an account without passing a project ID:

```go
account, err := client.Manager.CreateAccount(ctx, types.CreateAccountRequest{Name: "ACME"})
```

Import `github.com/meterry-com/meterry-go/pkg/types` for request and response types.

To issue a short-lived subject JWT, use a project-bound tenant API key with the
`subject_jwt:grant` scope (`types.TenantAPIKeyScopeSubjectJWTGrant`):

```go
grant, err := client.Manager.GrantSubject(ctx, types.GrantSubjectRequest{
	SubjectType: "user",
	SubjectID:   "user_123",
	Rules:       []string{"analysis"},
	Metrics:     []string{"input_tokens", "output_tokens"},
	ExpiresIn:   900,
})
if err != nil {
	return err
}
token := grant.AccessToken
```

This calls `POST /v1/subject/grants`. The server derives tenant and project
identity from the API key; use a key bound to the configured project. Fixed
service API keys cannot issue subject tokens. `Rules` currently supports
`analysis`. Omit `Metrics` to allow all metrics, or supply a nonempty list to
restrict access. `ExpiresIn` is in seconds; omit it to use the server's default
lifetime. The response includes `TokenType`, `AccessToken`, `ExpiresIn`,
`ExpiresAt` (Unix seconds), `Subject`, `Rules`, and optional `Metrics`.

This is a breaking API change: remove the `ForProject` suffix and the project ID argument from existing calls. Rule-set, report, event-listing, and ingest methods also use the configured project; `ChartQuery` no longer has a `ProjectID` field.

Metric-based analytics and bill queries should send explicit `Metrics` selectors, even for a single metric. The legacy singular `Metric` field on usage analytics requests is kept as a compatibility alias.

For webhook signature verification, use:

```go
err := sdk.VerifyWebhookRequest(req, body, secret)
```

## Wallet credit remarks

`CreditWallet` and its `TopUpWallet` alias accept an optional `Remark` to explain a credit. `SourceID` remains the identifier of the related business record.

```go
credit, err := client.Manager.CreditWallet(ctx, types.CreditWalletRequest{
    AccountID:      "acct_001",
    Currency:       "USD",
    Amount:         decimal.RequireFromString("100.00"),
    SourceType:     "manual_topup",
    SourceID:       "payment_001",
    Remark:         "Record a confirmed customer bank transfer",
    IdempotencyKey: "topup:payment_001",
})
```

The service trims leading and trailing whitespace and allows up to 500 Unicode characters. Remarks are visible to ledger readers; avoid internal sensitive information. An empty remark is omitted from JSON. Credit responses and `ListWalletLedgerEntries` expose the stored value as `WalletLedgerEntry.Remark`. A retry with the same idempotency key preserves the original remark. Legacy entries without this field decode with an empty remark.

`CreateVoucherGrant` also accepts `types.CreateVoucherGrantRequest.Remark`. The reason is stored only on the credit `WalletLedgerEntry.Remark` and can be read through the create response or `ListWalletLedgerEntries`. Voucher grant records and `ListVoucherGrants` contain funding state, without remarks. The same optional-field, 500-character, whitespace-trimming, visibility, and idempotency rules apply.

## Official docs

[Meterry Go SDK documentation](https://meterry.com/docs/sdk/go/)
