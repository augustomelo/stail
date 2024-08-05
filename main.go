package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

// API docs https://docs.datadoghq.com/api/latest/logs/?code-lang=go#get-a-list-of-logs
// API docs https://docs.datadoghq.com/logs/guide/collect-multiple-logs-with-pagination/?tab=v2api
// IO reader for the body https://yourbasic.org/golang/io-reader-interface-explained/
// request object https://github.com/DataDog/datadog-api-client-go/blob/8f7bfb291511b4fa689c2fc0c2826d75b731e3c8/api/datadogV2/api_logs.go#L268

// LogsStorageTier Specifies storage type as indexes, online-archives or flex

// The idea is to have a Producer which will grab the data from a API for example,
// and a conrumerwhich will grab the data from the Producer and tranform into
// something

type Integration string

const (
	DATADOG_API_KEY             string = "DD-API-KEY"
	DATADOG_APPLICATION_KEY     string = "DD-APPLICATION-KEY"
	DATADOG_ENV_API_KEY         string = "DD_API_KEY"
	DATADOG_ENV_APPLICATION_KEY string = "DD_APPLICATION_KEY"

	DATADOG_QUERY_FILTER_QUERY    string = "filter[query]"
	DATADOG_QUERY_PAGE_LIMIT      string = "page[limit]"
	DATADOG_QUERY_SORT            string = "sort"
	DATADOG_QUERY_SORT_ASCENDING  string = "timestamp"
	DATADOG_QUERY_SORT_DESCENDING string = "-timestamp"

	HTTP_HEADER_ACCEPT           string = "Accept"
	HTTP_HEADER_APPLICATION_JSON string = "application/json"

	DATADOG Integration = "DATADOG"
)

type Producer interface {
	Produce(context.Context) (string, error)
}

type DataDogSource struct {
	Headers http.Header
	URL     *url.URL
	Client  *http.Client
	Type    Integration
}

// Read from XDG ?
func buildDataDogSource() *DataDogSource {
	ddSource := &DataDogSource{
		Headers: http.Header{
			HTTP_HEADER_ACCEPT:      {HTTP_HEADER_APPLICATION_JSON},
			DATADOG_API_KEY:         {os.Getenv(DATADOG_ENV_API_KEY)},
			DATADOG_APPLICATION_KEY: {os.Getenv(DATADOG_ENV_APPLICATION_KEY)},
		},
		URL: &url.URL{
			Scheme: "https",
			Host:   "api.datadoghq.eu",
			Path:   "/api/v2/logs/events",
		},
		Client: &http.Client{
			Timeout: time.Duration(1) * time.Second,
		},
		Type: DATADOG,
	}

	slog.Debug("Built DataDogSource", "source", ddSource)

	return ddSource
}

func (source DataDogSource) Produce(context.Context) (string, error) {
	query := source.URL.Query()
	query.Set(DATADOG_QUERY_FILTER_QUERY, "")
	query.Set(DATADOG_QUERY_SORT, DATADOG_QUERY_SORT_ASCENDING)
	query.Set(DATADOG_QUERY_PAGE_LIMIT, "1")
	source.URL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, source.URL.String(), nil)
	if err != nil {
		slog.Error("Error while creating request", "err", err)
		return "", err
	}

	req.Header = source.Headers

	resp, err := source.Client.Do(req)
	if err != nil {
		slog.Error("Error while performing the request", "err", err)
		return "", err
	}

	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error while reading the response", "err", err)
		return "", err
	}

	slog.Debug("Body resposse", "body", bodyResponse)
	return string(bodyResponse), err
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Debug("Start")

	source := buildDataDogSource()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source.Produce(ctx)

	slog.Debug("End")
}
