package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/augustomelo/stail/internal/ui/view"
	tea "github.com/charmbracelet/bubbletea"
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

type DataDogLogResponse struct {
	Data []struct {
		Attributes struct {
			Attributes map[string]any `json:"attributes,omitempty"`
			Host       string         `json:"host,omitempty"`
			Message    string         `json:"message,omitempty"`
			Service    string         `json:"service,omitempty"`
			Status     string         `json:"status,omitempty"`
			Tags       []string       `json:"tags,omitempty"`
			Timestamp  time.Time      `json:"timestamp"`
		} `json:"attributes,omitempty"`
		ID   string `json:"id"`
		Type string `json:"type,omitempty"`
	} `json:"data,omitempty"`
	Links struct {
		Next string `json:"next,omitempty"`
	} `json:"links,omitempty"`
	Meta struct {
		Elapsed int `json:"elapsed,omitempty"`
		Page    struct {
			After string `json:"after,omitempty"`
		} `json:"page,omitempty"`
		RequestID string `json:"request_id,omitempty"`
		Status    string `json:"status,omitempty"`
		Warnings  []struct {
			Code   string `json:"code,omitempty"`
			Detail string `json:"detail,omitempty"`
			Title  string `json:"title,omitempty"`
		} `json:"warnings,omitempty"`
	} `json:"meta,omitempty"`
}

type Producer interface {
	Produce(context.Context) (*[]byte, error)
}

type Mapper interface {
	Map(*[]byte)
}

type Log struct {
	ID         string
	Timestamp  time.Time
	Attributes map[string]any
	Level      string
	Message    string
	Tags       []string
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

func (source DataDogSource) Produce(ctx context.Context) (*[]byte, error) {
	query := source.URL.Query()
	query.Set(DATADOG_QUERY_FILTER_QUERY, "")
	query.Set(DATADOG_QUERY_SORT, DATADOG_QUERY_SORT_ASCENDING)
	query.Set(DATADOG_QUERY_PAGE_LIMIT, "1")
	source.URL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, source.URL.String(), nil)
	if err != nil {
		slog.Error("Error while creating request", "err", err)
		return nil, err
	}

	req.Header = source.Headers

	resp, err := source.Client.Do(req)
	if err != nil {
		slog.Error("Error while performing the request", "err", err)
		return nil, err
	}

	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error while reading the response", "err", err)
		return nil, err
	}

	// Should be dealing the erros 400, 403, 429
	// https://docs.datadoghq.com/api/latest/logs/?code-lang=go#get-a-list-of-logs
	slog.Debug("Body resposse", "body", string(bodyResponse))
	return &bodyResponse, err
}

func (source DataDogSource) Map(body *[]byte) {
	var ddResponse DataDogLogResponse

	err := json.Unmarshal(*body, &ddResponse)
	if err != nil {
		slog.Error("Error while unmarshalling the response", "err", err)
	}

	var logs []Log
	for _, v := range ddResponse.Data {
		logs = append(logs, Log{
			Attributes: v.Attributes.Attributes,
			ID:         v.ID,
			Level:      v.Attributes.Status,
			Message:    v.Attributes.Message,
			Tags:       v.Attributes.Tags,
			Timestamp:  v.Attributes.Timestamp,
		})
	}

	slog.Debug("Mapped logs", "logs", logs)
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	slog.Debug("Start")

	p := tea.NewProgram(
		view.InitialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		slog.Error("Error while instantiating view", "err", err)
		os.Exit(1)
	}

	source := buildDataDogSource()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	body, err := source.Produce(ctx)
	if err != nil {
		slog.Error("Error while producing logs", "err", err)
	}
	source.Map(body)

	slog.Debug("End")
}
