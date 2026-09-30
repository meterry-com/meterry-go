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

No tenant ID is needed at initialization. Tenant identity is authenticated by the API key; the SDK does not send a tenant header. All operations use the configured project automatically. Create a separate client for each tenant/project pair.

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
This is a breaking API change: remove the `ForProject` suffix and the project ID argument from existing calls. Rule-set, report, event-listing, and ingest methods also use the configured project; `ChartQuery` no longer has a `ProjectID` field.

Metric-based analytics and bill queries should send explicit `Metrics` selectors, even for a single metric. The legacy singular `Metric` field on usage analytics requests is kept as a compatibility alias.

For webhook signature verification, use:

```go
err := sdk.VerifyWebhookRequest(req, body, secret)
```

## Official docs

[Meterry Go SDK documentation](https://meterry.com/docs/sdk/go/)
