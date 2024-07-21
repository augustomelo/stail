package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// API docs https://docs.datadoghq.com/api/latest/logs/?code-lang=go#get-a-list-of-logs
// API docs https://docs.datadoghq.com/logs/guide/collect-multiple-logs-with-pagination/?tab=v2api
// IO reader for the body https://yourbasic.org/golang/io-reader-interface-explained/
// request object https://github.com/DataDog/datadog-api-client-go/blob/8f7bfb291511b4fa689c2fc0c2826d75b731e3c8/api/datadogV2/api_logs.go#L268

// LogsStorageTier Specifies storage type as indexes, online-archives or flex
const (
	DATADOG_TIME_ASCENDING      string = "timestamp"
	DATADOG_TIME_DESCENDING     string = "-timestamp"
	DATADOG_API_KEY             string = "DD-API-KEY"
	DATADOG_APPLICATION_KEY     string = "DD-APPLICATION-KEY"
	DATADOG_ENV_API_KEY         string = "DD_API_KEY"
	DATADOG_ENV_APPLICATION_KEY string = "DD_APPLICATION_KEY"

	HTTP_HEADER_ACCEPT string = "Accept"
)

func main() {
	fmt.Println("Start")
	fetchLogs()
	fmt.Println("End")
}

func fetchLogs() {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://api.datadoghq.eu/api/v2/logs/events", nil)
	if err != nil {
		fmt.Print(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(DATADOG_API_KEY, os.Getenv(DATADOG_ENV_API_KEY))
	req.Header.Add(DATADOG_APPLICATION_KEY, os.Getenv(DATADOG_ENV_APPLICATION_KEY))

	req.URL.RawQuery = url.Values{
		"filter[query]": []string{""},
		"sort":          []string{DATADOG_TIME_ASCENDING},
		"page[limit]":   []string{"1"},
	}.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
		return
	}

	defer resp.Body.Close()
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println(string(bodyResponse))
}
