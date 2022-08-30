package druid

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/apache/calcite-avatica-go/v5"
	"github.com/rilldata/rill/runtime/infra"
)

func init() {
	infra.Register("druid", driver{})
}

const requestTimeout = 5 * time.Minute

type driver struct{}

// Using Avatica driver for druid
func (d driver) Open(dsn string) (infra.Connection, error) {
	db, err := sqlx.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{db: db}
	return conn, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

func (c *connection) InformationSchema() string {
	return ""
}

func (c *connection) Execute(ctx context.Context, priority int, sql string, args ...any) (*sqlx.Rows, error) {
	rows, err := c.db.QueryxContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type TaskResult struct {
	Task string
}

type IngestionTaskReport struct {
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

type DatasourceDetails struct {
	Segments struct {
		Count float64
	}
}

func Ingest(druidCoordinatorUrl string, indexJsonStr string, dataSourceName string) (*http.Response, error) {
	reader := strings.NewReader(indexJsonStr)
	taskUrl := fmt.Sprintf("%s/druid/indexer/v1/task", druidCoordinatorUrl)

	req, err := http.NewRequest("POST", taskUrl, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status TaskResult
	err = json.Unmarshal(body, &status)
	if err != nil {
		return nil, err
	}

	deadline := time.Now().Add(requestTimeout)

	for {
		time.Sleep(2 * time.Second)

		ingestionTaskReport, err := getTaskReport(druidCoordinatorUrl, status.Task)
		if err != nil {
			return nil, err
		}

		datasourceDetails, err := getDatasourceDetails(druidCoordinatorUrl, dataSourceName)
		if err != nil {
			return nil, err
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("ingestion timeout")
		}

		success := ingestionTaskReport.IngestionStatsAndErrors.Payload.IngestionState != ""
		if !success {
			continue
		}

		segmentsProcessedCount := ingestionTaskReport.IngestionStatsAndErrors.Payload.RowStats.BuildSegments.Processed
		ingestionState := ingestionTaskReport.IngestionStatsAndErrors.Payload.IngestionState
		segmentsSubmittedCount := datasourceDetails.Segments.Count

		if ingestionState == "COMPLETED" && segmentsProcessedCount == segmentsSubmittedCount {
			return resp, nil
		}

	}
}

func getDatasourceDetails(druidUrl string, dataSourceName string) (*DatasourceDetails, error) {
	dataSourceUrl := fmt.Sprintf("%s/druid/coordinator/v1/datasources/%s", druidUrl, dataSourceName)
	method := "GET"

	var datasourceDetails DatasourceDetails
	err := sendRequest(method, dataSourceUrl, &datasourceDetails)
	if err != nil {
		return nil, err
	}

	return &datasourceDetails, err
}

func getTaskReport(druidUrl string, taskId string) (*IngestionTaskReport, error) {
	taskReportUrl := fmt.Sprintf("%s/druid/indexer/v1/task/%s/reports", druidUrl, taskId)
	method := "GET"

	var ingestionTaskReport IngestionTaskReport
	err := sendRequest(method, taskReportUrl, &ingestionTaskReport)
	if err != nil {
		return nil, err
	}

	return &ingestionTaskReport, err
}

func sendRequest(method string, url string, out any) error {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.StatusCode != 200 {
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return err
	}

	return nil
}
