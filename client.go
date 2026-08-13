package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const defaultUserAgent = "edgefn-billing-go-sdk"

type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

type Client struct {
	Manager *ManagerClient
	Query   *QueryClient
	Ingest  *IngestClient

	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

func NewClient(config Config) (*Client, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		return nil, errors.New("billing sdk base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse billing sdk base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("billing sdk base url must include scheme and host")
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	client := &Client{
		baseURL:    parsed,
		apiKey:     strings.TrimSpace(config.APIKey),
		httpClient: httpClient,
		userAgent:  firstNonEmpty(config.UserAgent, defaultUserAgent),
	}
	client.Manager = &ManagerClient{client: client}
	client.Query = &QueryClient{client: client}
	client.Ingest = &IngestClient{client: client}
	return client, nil
}

func MustNewClient(config Config) *Client {
	client, err := NewClient(config)
	if err != nil {
		panic(err)
	}
	return client
}

func (c *Client) request(ctx context.Context, method, endpoint string, query url.Values, body any, out any) error {
	if c == nil {
		return errors.New("billing sdk client is nil")
	}
	reqURL := *c.baseURL
	reqURL.Path = joinURLPath(c.baseURL.Path, endpoint)
	reqURL.RawQuery = query.Encode()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal billing request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL.String(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return decodeAPIError(resp.StatusCode, raw)
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(out); err != nil {
		return fmt.Errorf("decode billing response: %w", err)
	}
	return nil
}

func (c *Client) apiPath(parts ...string) string {
	return joinURLPath("/v1", path.Join(parts...))
}

func joinURLPath(base string, parts ...string) string {
	values := make([]string, 0, len(parts)+1)
	if strings.TrimSpace(base) != "" {
		values = append(values, strings.TrimRight(base, "/"))
	}
	for _, part := range parts {
		part = strings.Trim(part, "/")
		if part != "" {
			values = append(values, part)
		}
	}
	if len(values) == 0 {
		return "/"
	}
	joined := path.Join(values...)
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("billing api error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("billing api error %d", e.StatusCode)
}

func decodeAPIError(statusCode int, raw []byte) error {
	apiErr := &APIError{StatusCode: statusCode, Body: string(raw)}
	var payload struct {
		Error string `json:"error"`
	}
	if len(raw) > 0 && json.Unmarshal(raw, &payload) == nil {
		apiErr.Message = payload.Error
	}
	if apiErr.Message == "" {
		apiErr.Message = strings.TrimSpace(string(raw))
	}
	return apiErr
}
