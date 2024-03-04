package datadog

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

type Client struct {
	api    *datadogV2.MetricsApi
	apiKey string
	appKey string
}

func NewDatadogClient(apiKey, appKey string) Client {
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewMetricsApi(apiClient)

	return Client{
		api:    api,
		apiKey: apiKey,
		appKey: appKey,
	}
}

func (c *Client) PublishMetric(ctx context.Context, metricName string) error {
	valueCtx := context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: c.apiKey,
			},
			"appKeyAuth": {
				Key: c.appKey,
			},
		},
	)

	body := datadogV2.MetricPayload{
		Series: []datadogV2.MetricSeries{
			{
				Metric: metricName,
				Type:   datadogV2.METRICINTAKETYPE_COUNT.Ptr(),
				Points: []datadogV2.MetricPoint{
					{
						Timestamp: datadog.PtrInt64(time.Now().Unix()),
						Value:     datadog.PtrFloat64(1),
					},
				},
				Resources: []datadogV2.MetricResource{
					{
						Type: datadog.PtrString("source"),
						Name: datadog.PtrString("pi-brew"),
					},
					{
						Type: datadog.PtrString("service"),
						Name: datadog.PtrString("pi-brew"),
					},
				},
			},
		},
	}

	_, _, err := c.api.SubmitMetrics(valueCtx, body, *datadogV2.NewSubmitMetricsOptionalParameters())
	if err != nil {
		return fmt.Errorf("submitting metrics: %s", err)
	}

	return nil
}
