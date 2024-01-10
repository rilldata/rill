package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	force "github.com/ForceCLI/force/lib"
	"github.com/ForceCLI/force/lib/record_reader"
	"go.uber.org/zap"
)

type batchResult struct {
	batch    *force.BatchInfo
	resultID string
}

type bulkJob struct {
	session    *force.Force
	objectName string
	query      string
	job        force.JobInfo
	jobID      string
	batchID    string
	logger     *zap.Logger
	// pkChunking automatically splits large data sets into smaller batches of pkChunkSize, which we can query concurrently later on
	pkChunkSize  int
	results      []batchResult
	nextResult   int
	tempFilePath string
}

func (j *bulkJob) RecordReader(in io.Reader) record_reader.RecordReader {
	return record_reader.NewCsv(in, &record_reader.Options{GroupSize: 100})
}

func makeBulkJob(session *force.Force, objectName, query string, queryAll bool, logger *zap.Logger) *bulkJob {
	pkChunkSize := 100000
	contentType := force.JobContentTypeCsv
	operation := "query"

	if queryAll {
		operation = "queryAll"
	}

	return &bulkJob{
		session:     session,
		objectName:  objectName,
		query:       query,
		pkChunkSize: pkChunkSize,
		logger:      logger,
		job: force.JobInfo{
			Operation:   operation,
			Object:      objectName,
			ContentType: string(contentType),
		},
	}
}

func (c *connection) startJob(ctx context.Context, j *bulkJob) error {
	session := j.session

	jobInfo, err := session.CreateBulkJobWithContext(ctx, j.job, func(request *http.Request) {
		if isPKChunkingEnabled(j) {
			pkChunkHeader := "chunkSize=" + strconv.Itoa(j.pkChunkSize)
			parent := parentObject(j.objectName)

			if len(parent) > 0 {
				pkChunkHeader += "; parent=" + parent
			}

			request.Header.Add("Sforce-Enable-PKChunking", pkChunkHeader)
		}
	})
	if err != nil {
		if errors.Is(err, force.InvalidBulkObject) {
			return errors.New("object is not supported by Bulk API")
		}
		return err
	}
	result, err := session.BulkQueryWithContext(ctx, j.query, jobInfo.Id, j.job.ContentType)
	if err != nil {
		return errors.New("bulk query failed with " + err.Error())
	}
	batchID := result.Id
	// wait for chunking to complete
	if isPKChunkingEnabled(j) {
		for {
			batchInfo, err := session.GetBatchInfoWithContext(ctx, jobInfo.Id, batchID)
			if err != nil {
				return errors.New("bulk job status failed with " + err.Error())
			}

			if batchInfo.State == "NotProcessed" {
				break
			}
			c.logger.Info("Waiting for pk chunking to complete")
			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				return fmt.Errorf("startJob cancelled: %w", ctx.Err())
			}
		}
	}

	jobInfo, err = session.CloseBulkJobWithContext(ctx, jobInfo.Id)
	if err != nil {
		return err
	}
	var status force.JobInfo

	for {
		status, err = session.GetJobInfoWithContext(ctx, jobInfo.Id)
		if err != nil {
			return errors.New("bulk job status failed with " + err.Error())
		}
		if status.NumberBatchesCompleted+status.NumberBatchesFailed == status.NumberBatchesTotal {
			break
		}
		c.logger.Info("Waiting for bulk export to complete")
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return fmt.Errorf("startJob cancelled: %w", ctx.Err())
		}
	}

	j.job = status
	j.jobID = jobInfo.Id
	j.batchID = batchID

	return nil
}

func (j *bulkJob) getBatches(ctx context.Context) error {
	if j.jobID == "" {
		return fmt.Errorf("Invalid job: no job id")
	}

	var batches []force.BatchInfo
	var err error
	errorMessage := "Could not retrieve job result. Reason: "

	if isPKChunkingEnabled(j) {
		var allBatches []force.BatchInfo
		allBatches, err = j.session.GetBatchesWithContext(ctx, j.job.Id)
		// for pk chunking enabled jobs the first batch has no results
		if allBatches != nil {
			if allBatches[0].State == "Failed" {
				return fmt.Errorf("Batch failed with: %s", allBatches[0].StateMessage)
			}

			for _, b := range allBatches {
				if b.State != "NotProcessed" && b.NumberRecordsProcessed > 0 {
					batches = append(batches, b)
				}
			}
		}
	} else {
		batch, berr := j.session.GetBatchInfoWithContext(ctx, j.jobID, j.batchID)
		err = berr
		batches = []force.BatchInfo{batch}
	}
	if err != nil {
		return fmt.Errorf("%s %w", errorMessage+"batch status failed with ", err)
	}
	for _, b := range batches {
		results, err := getBatchResults(ctx, j.session, j.job, b)
		if err != nil {
			return fmt.Errorf("%s %w", errorMessage+"batch results failed with ", err)
		}
		j.results = append(j.results, results...)
	}
	return nil
}

func (j *bulkJob) retrieveJobResult(ctx context.Context, result int) (string, error) {
	batchResult := j.results[result]
	writer, err := os.CreateTemp("", "batchResult-"+batchResult.resultID+"-*.csv")
	if err != nil {
		return "", err
	}
	defer func() {
		writer.Close()
	}()

	httpBody := fetchBatchResult(ctx, j, batchResult, j.logger)
	err = readAndWriteBody(ctx, j, httpBody, writer)
	if closer, ok := httpBody.(io.ReadCloser); ok {
		closer.Close()
	}
	if err != nil {
		return "", err
	}
	return writer.Name(), nil
}

func fetchBatchResult(ctx context.Context, j *bulkJob, resultInfo batchResult, logger *zap.Logger) io.Reader {
	errorMessage := "Could not fetch job result. Reason: "

	if resultInfo.batch.State == "Failed" {
		logger.Error(errorMessage + "batch failed with " + resultInfo.batch.StateMessage)
		return bytes.NewReader(nil)
	}
	if resultInfo.batch.NumberRecordsProcessed == 0 {
		logger.Debug("No records found for query")
		return bytes.NewReader(nil)
	}
	var result io.Reader
	err := j.session.RetrieveBulkJobQueryResultsWithCallbackWithContext(ctx, j.job, resultInfo.batch.Id, resultInfo.resultID, func(r *http.Response) error {
		result = r.Body
		return nil
	})
	if err != nil {
		logger.Error(errorMessage + "batch failed with " + err.Error())
		return bytes.NewReader(nil)
	}
	return result
}

func readAndWriteBody(ctx context.Context, j *bulkJob, httpBody io.Reader, w io.Writer) error {
	recReader := j.RecordReader(httpBody)
	for {
		records, err := recReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
		if _, err := io.Copy(w, bytes.NewReader(records.Bytes)); err != nil {
			return fmt.Errorf("write failed: %w", err)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("readAndWriteBody cancelled: %w", ctx.Err())
		default:
		}
	}
}

// Get all of the results for a batch.  Most batches have one results, but
// large batches can be split into multiple result files.
func getBatchResults(ctx context.Context, session *force.Force, job force.JobInfo, batch force.BatchInfo) ([]batchResult, error) {
	var resultIDs []string
	var results []batchResult
	jobInfo, err := session.RetrieveBulkQueryWithContext(ctx, job.Id, batch.Id)
	if err != nil {
		return nil, err
	}

	jct, err := job.JobContentType()
	if err != nil {
		return nil, err
	}
	if jct == force.JobContentTypeJson {
		err = json.Unmarshal(jobInfo, &resultIDs)
	} else {
		var resultList struct {
			Results []string `xml:"result"`
		}
		err = xml.Unmarshal(jobInfo, &resultList)
		resultIDs = resultList.Results
	}
	if err != nil {
		return nil, err
	}
	for _, r := range resultIDs {
		results = append(results, batchResult{batch: &batch, resultID: r})
	}

	return results, err
}
