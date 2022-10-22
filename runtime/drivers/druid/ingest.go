package druid

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type taskResult struct {
	Task string
}

type taskReport struct {
	IngestionStatsAndErrors struct {
		Payload struct {
			RowStats struct {
				BuildSegments struct {
					Processed float64
				}
			}
			IngestionState string
		}
	}
}

type datasourceDetails struct {
	Segments struct {
		Count float64
	}
}

// Ingest uses the Druid REST API to submit an ingestion spec. It returns once Druid has finished ingesting data.
// This function is for test and development usage and has not been tested for production use.
func Ingest(coordinatorURL string, specJSON string, datasourceName string, timeout time.Duration) error {
	var status taskResult
	err := sendRequest(coordinatorURL, http.MethodPost, "/druid/indexer/v1/task", specJSON, &status)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for {
		time.Sleep(pollInterval)

		if time.Now().After(deadline) {
			return fmt.Errorf("ingestion timeout")
		}

		tr, err := getTaskReport(coordinatorURL, status.Task)
		if err != nil {
			// The coordinator may return 404 or 500 on the first few polls
			if strings.Contains(err.Error(), "failed with status:") {
				continue
			}
			return err
		}

		success := tr.IngestionStatsAndErrors.Payload.IngestionState != ""
		if !success {
			continue
		}

		ds, err := getDatasourceDetails(coordinatorURL, datasourceName)
		if err != nil {
			return err
		}

		segmentsProcessedCount := tr.IngestionStatsAndErrors.Payload.RowStats.BuildSegments.Processed
		ingestionState := tr.IngestionStatsAndErrors.Payload.IngestionState
		segmentsSubmittedCount := ds.Segments.Count

		if ingestionState == "COMPLETED" && segmentsProcessedCount == segmentsSubmittedCount {
			return nil
		}
	}
}

func getTaskReport(coordinatorURL string, taskId string) (*taskReport, error) {
	var res taskReport
	path := fmt.Sprintf("/druid/indexer/v1/task/%s/reports", taskId)
	err := sendRequest(coordinatorURL, http.MethodGet, path, "", &res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func getDatasourceDetails(coordinatorURL string, datasourceName string) (*datasourceDetails, error) {
	var res datasourceDetails
	path := fmt.Sprintf("/druid/coordinator/v1/datasources/%s", datasourceName)
	err := sendRequest(coordinatorURL, http.MethodGet, path, "", &res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func sendRequest(coordinatorURL string, method string, path string, jsonBody string, out any) error {
	url, err := url.JoinPath(coordinatorURL, path)
	if err != nil {
		return err
	}

	var reqBody io.Reader
	if jsonBody != "" {
		reqBody = strings.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("coordinator request failed with status: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, out)
		if err != nil {
			return err
		}
	}

	return nil
}
