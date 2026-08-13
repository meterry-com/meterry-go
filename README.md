# meterry-go

Go SDK for Meterry.

## Install

```bash
go get github.com/meterry-com/meterry-go
```

## Setup

Create a client with your Meterry API base URL and API key:

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
		BaseURL: os.Getenv("METERRY_BASE_URL"),
		APIKey:  os.Getenv("METERRY_API_KEY"),
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
- `APIKey`: optional bearer token
- `HTTPClient`: optional custom HTTP client
- `UserAgent`: optional custom user agent

If you prefer a panic-on-error constructor, use `sdk.MustNewClient(...)`.

## Usage

The client exposes three main entry points:

- `client.Manager`
- `client.Query`
- `client.Ingest`

For webhook signature verification, use:

```go
err := sdk.VerifyWebhookRequest(req, body, secret)
```

## Official docs

[Meterry Go SDK documentation](https://meterry.com/docs/sdk/go/)
