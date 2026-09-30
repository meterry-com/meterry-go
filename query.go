package sdk

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/meterry-com/meterry-go/pkg/types"
)

type QueryClient struct {
	client *Client
}

type ChartQuery struct {
	// Metric is a backward-compatible single-metric alias.
	Metric string
	// Metrics is the preferred explicit selector list.
	Metrics     []string
	Measure     string
	BucketSize  string
	StartTime   int64
	EndTime     int64
	SubjectType string
	SubjectID   string
	Filters     map[string]string
	Limit       int
}

type MetricSeries struct {
	Metric string       `json:"metric,omitempty"`
	Points []ChartPoint `json:"points"`
}

type ChartPoint struct {
	Time  int64           `json:"time"`
	Value decimal.Decimal `json:"value"`
}

func (q *QueryClient) Usage(ctx context.Context, req types.UsageAnalyticsQueryRequest) (*types.UsageAnalyticsQueryResponse, error) {
	var err error
	if req, err = normalizeAnalyticsQueryRequest(req); err != nil {
		return nil, err
	}
	var out types.UsageAnalyticsQueryResponse
	err = q.client.request(ctx, http.MethodPost, q.client.projectPath("usage", "query"), nil, req, &out)
	return &out, err
}

func (q *QueryClient) Bill(ctx context.Context, req types.UsageBillQueryRequest) (*types.UsageBillQueryResponse, error) {
	var err error
	if req, err = normalizeBillQueryRequest(req); err != nil {
		return nil, err
	}
	var out types.UsageBillQueryResponse
	err = q.client.request(ctx, http.MethodPost, q.client.projectPath("usage", "bills", "query"), nil, req, &out)
	return &out, err
}

func (q *QueryClient) Report(ctx context.Context, reportID string, req types.UsageAnalyticsQueryRequest) (*types.UsageAnalyticsQueryResponse, error) {
	var err error
	if req, err = normalizeAnalyticsQueryRequest(req); err != nil {
		return nil, err
	}
	var out types.UsageAnalyticsQueryResponse
	err = q.client.request(ctx, http.MethodPost, q.client.projectPath("reports", reportID, "query"), nil, req, &out)
	return &out, err
}

func (q *QueryClient) Chart(ctx context.Context, req ChartQuery) ([]MetricSeries, error) {
	var err error
	if req, err = normalizeChartQuery(req); err != nil {
		return nil, err
	}
	if len(req.Metrics) > 1 {
		series := make([]MetricSeries, 0, len(req.Metrics))
		for _, metric := range req.Metrics {
			next := req
			next.Metric = metric
			next.Metrics = nil
			metricSeries, err := q.chartSingleMetric(ctx, next)
			if err != nil {
				return nil, err
			}
			if len(metricSeries) == 0 {
				series = append(series, MetricSeries{Metric: metric})
				continue
			}
			metricSeries[0].Metric = metric
			series = append(series, metricSeries[0])
		}
		return series, nil
	}
	return q.chartSingleMetric(ctx, req)
}

func (q *QueryClient) chartSingleMetric(ctx context.Context, req ChartQuery) ([]MetricSeries, error) {
	measure := firstNonEmpty(req.Measure, "quantity")
	bucketSize := firstNonEmpty(req.BucketSize, "5m")
	groupBy := []string{"bucket"}
	analyticsReq := types.UsageAnalyticsQueryRequest{
		BucketSize:  bucketSize,
		SubjectType: req.SubjectType,
		SubjectID:   req.SubjectID,
		Metric:      req.Metric,
		Metrics:     req.Metrics,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Filters:     req.Filters,
		GroupBy:     groupBy,
		Measures:    []string{measure},
		Limit:       req.Limit,
	}
	resp, err := q.Usage(ctx, analyticsReq)
	if err != nil {
		return nil, err
	}
	return chartSeriesFromRows(resp.Rows, measure), nil
}

func (q *QueryClient) ListUsageLogs(ctx context.Context, req types.ListUsageLogsRequest) (*types.ListUsageLogsResponse, error) {
	query := usageLogQuery(req)
	var out types.ListUsageLogsResponse
	err := q.client.request(ctx, http.MethodGet, q.client.projectPath("usage-logs"), query, nil, &out)
	return &out, err
}

func (q *QueryClient) ListProjectUsageEvents(ctx context.Context, req types.ListUsageEventLogsRequest) (*types.ListUsageEventLogsResponse, error) {
	query := usageEventLogQuery(req)
	var out types.ListUsageEventLogsResponse
	err := q.client.request(ctx, http.MethodGet, q.client.projectPath("usage", "events"), query, nil, &out)
	return &out, err
}

func (q *QueryClient) ListProjectUsageEventDetails(ctx context.Context, req types.ListUsageEventLogsRequest) (*types.ListUsageEventDetailsResponse, error) {
	query := usageEventLogQuery(req)
	var out types.ListUsageEventDetailsResponse
	err := q.client.request(ctx, http.MethodGet, q.client.projectPath("usage", "events", "details"), query, nil, &out)
	return &out, err
}

func chartSeriesFromRows(rows []types.UsageAnalyticsRow, measure string) []MetricSeries {
	byMetric := map[string][]ChartPoint{}
	for _, row := range rows {
		bucket := row.Dimensions["bucket"]
		if bucket == "" {
			continue
		}
		ts, err := strconv.ParseInt(bucket, 10, 64)
		if err != nil {
			continue
		}
		metric := row.Dimensions["metric"]
		byMetric[metric] = append(byMetric[metric], ChartPoint{
			Time:  ts,
			Value: row.Measures[measure],
		})
	}
	metrics := make([]string, 0, len(byMetric))
	for metric := range byMetric {
		metrics = append(metrics, metric)
	}
	sort.Strings(metrics)
	series := make([]MetricSeries, 0, len(metrics))
	for _, metric := range metrics {
		points := byMetric[metric]
		sort.Slice(points, func(i, j int) bool {
			return points[i].Time < points[j].Time
		})
		series = append(series, MetricSeries{Metric: metric, Points: points})
	}
	return series
}

func normalizeAnalyticsQueryRequest(req types.UsageAnalyticsQueryRequest) (types.UsageAnalyticsQueryRequest, error) {
	metrics := normalizeMetricsSelector(req.Metric, req.Metrics)
	if requiresMetricSelector(req.Measures) && len(metrics) == 0 {
		return req, errors.New("metric or metrics is required for metric measures")
	}
	req.Metric = ""
	req.Metrics = metrics
	return req, nil
}

func normalizeBillQueryRequest(req types.UsageBillQueryRequest) (types.UsageBillQueryRequest, error) {
	req.Metrics = normalizeMetricsSelector("", req.Metrics)
	if len(req.Metrics) == 0 {
		return req, errors.New("metrics are required for bill queries")
	}
	return req, nil
}

func normalizeChartQuery(req ChartQuery) (ChartQuery, error) {
	req.Metrics = normalizeMetricsSelector(req.Metric, req.Metrics)
	req.Metric = ""
	if len(req.Metrics) == 0 {
		return req, errors.New("metrics are required for chart queries")
	}
	return req, nil
}

func normalizeMetricsSelector(metric string, metrics []string) []string {
	if len(metrics) > 0 {
		result := make([]string, 0, len(metrics))
		seen := make(map[string]struct{}, len(metrics))
		for _, value := range metrics {
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
		return result
	}
	if metric == "" {
		return nil
	}
	return []string{metric}
}

func requiresMetricSelector(measures []string) bool {
	for _, measure := range measures {
		switch measure {
		case "usage_event_count", "event_count":
		default:
			return true
		}
	}
	return false
}

func usageLogQuery(req types.ListUsageLogsRequest) url.Values {
	query := url.Values{}
	addQuery(query, "usage_event_id", req.UsageEventID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	addQuery(query, "metric", req.Metric)
	for _, metric := range req.Metrics {
		addQuery(query, "metrics", metric)
	}
	addInt64Query(query, "start_time", req.StartTime)
	addInt64Query(query, "end_time", req.EndTime)
	addIntQuery(query, "limit", req.Limit)
	addFilterQuery(query, req.Filters)
	return query
}

func usageEventLogQuery(req types.ListUsageEventLogsRequest) url.Values {
	query := url.Values{}
	addQuery(query, "usage_event_id", req.UsageEventID)
	addQuery(query, "raw_event_id", req.RawEventID)
	addQuery(query, "subject_type", req.SubjectType)
	addQuery(query, "subject_id", req.SubjectID)
	addQuery(query, "metric", req.Metric)
	for _, metric := range req.Metrics {
		addQuery(query, "metrics", metric)
	}
	addInt64Query(query, "start_time", req.StartTime)
	addInt64Query(query, "end_time", req.EndTime)
	addQuery(query, "after_id", req.AfterID)
	addIntQuery(query, "limit", req.Limit)
	addFilterQuery(query, req.Filters)
	return query
}

func addFilterQuery(query url.Values, filters map[string]string) {
	for key, value := range filters {
		addQuery(query, "filter."+key, value)
	}
}

func addQuery(query url.Values, key, value string) {
	if value != "" {
		query.Add(key, value)
	}
}

func addInt64Query(query url.Values, key string, value int64) {
	if value != 0 {
		query.Add(key, strconv.FormatInt(value, 10))
	}
}

func addIntQuery(query url.Values, key string, value int) {
	if value != 0 {
		query.Add(key, strconv.Itoa(value))
	}
}
