package salesforce

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	force "github.com/ForceCLI/force/lib"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

// bulk2QueryJob drives a Salesforce Bulk API 2.0 query job and exposes the
// paginated CSV results through drivers.FileIterator. Each Next() call fetches
// the next page (one CSV file per page) using the Sforce-Locator cursor.
type bulk2QueryJob struct {
	session *force.Force
	logger  *zap.Logger

	jobID               string
	locator             string
	done                bool
	tempFilePaths       []string
	keepFilesUntilClose bool
}

const bulk2PollInterval = 2 * time.Second

func makeBulk2QueryJob(session *force.Force, logger *zap.Logger) *bulk2QueryJob {
	return &bulk2QueryJob{session: session, logger: logger}
}

// startJob submits the query to Bulk API 2.0 and waits for the job to reach a
// terminal state before returning. The caller can then iterate results via Next().
func (j *bulk2QueryJob) startJob(ctx context.Context, query string, queryAll bool) error {
	op := force.Bulk2OperationQuery
	if queryAll {
		op = force.Bulk2OperationQueryAll
	}

	// The connector explorer's "Table" mode (and users typing the same
	// shape) produces `SELECT * FROM <SObject>`, which SOQL doesn't accept.
	// Expand the star into the SObject's queryable field list before sending.
	expanded, err := expandSelectStar(j.session, query)
	if err != nil {
		return err
	}
	query = expanded

	info, err := j.session.CreateBulk2QueryJobWithContext(ctx, force.Bulk2QueryJobRequest{
		Operation: op,
		Query:     query,
	})
	if err != nil {
		return fmt.Errorf("creating Bulk API 2.0 query job: %w", err)
	}
	j.jobID = info.Id

	final, err := j.session.WaitForBulk2QueryJobWithContext(ctx, j.jobID, bulk2PollInterval, func(state any) {
		info, _ := state.(force.Bulk2QueryJobInfo)
		j.logger.Info("Waiting for Bulk API 2.0 query job", zap.String("state", string(info.State)), observability.ZapCtx(ctx))
	})
	if err != nil {
		return fmt.Errorf("waiting for Bulk API 2.0 query job: %w", err)
	}
	if final.State == force.Bulk2JobStateFailed {
		msg := final.ErrorMessage
		if msg == "" {
			msg = "job failed without an error message"
		}
		return fmt.Errorf("Bulk API 2.0 query job failed: %s", msg)
	}
	if final.State == force.Bulk2JobStateAborted {
		return errors.New("Bulk API 2.0 query job was aborted")
	}
	if final.NumberRecordsProcessed == 0 {
		// Nothing to download; mark done so Next() returns io.EOF immediately.
		j.done = true
	}
	return nil
}

var _ drivers.FileIterator = &bulk2QueryJob{}

// Close implements drivers.FileIterator.
func (j *bulk2QueryJob) Close() error {
	return j.cleanupTempFiles()
}

// Format implements drivers.FileIterator.
func (j *bulk2QueryJob) Format() string { return "csv" }

// SetKeepFilesUntilClose implements drivers.FileIterator.
func (j *bulk2QueryJob) SetKeepFilesUntilClose() { j.keepFilesUntilClose = true }

// Next implements drivers.FileIterator. Each call fetches one page of results
// using the locator cursor returned by the previous call.
func (j *bulk2QueryJob) Next(ctx context.Context) ([]string, error) {
	if j.jobID == "" {
		return nil, errors.New("invalid bulk2 job: no job id")
	}
	if j.done {
		return nil, io.EOF
	}

	if !j.keepFilesUntilClose {
		if err := j.cleanupTempFiles(); err != nil {
			return nil, err
		}
	}

	page, err := j.session.GetBulk2QueryResultsWithContext(ctx, j.jobID, j.locator, 0)
	if err != nil {
		return nil, fmt.Errorf("fetching Bulk API 2.0 results: %w", err)
	}
	j.locator = page.Locator
	if j.locator == "" {
		// Salesforce signals the final page by returning an empty Sforce-Locator
		// header; mark done so the next iteration returns io.EOF.
		j.done = true
	}

	path, err := writeBytesToTempFile(page.Data, j.jobID)
	if err != nil {
		return nil, err
	}
	j.tempFilePaths = append(j.tempFilePaths, path)
	return []string{path}, nil
}

func (j *bulk2QueryJob) cleanupTempFiles() error {
	if len(j.tempFilePaths) == 0 {
		return nil
	}
	for _, p := range j.tempFilePaths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete temp file %s: %w", p, err)
		}
	}
	j.tempFilePaths = nil
	return nil
}

func writeBytesToTempFile(data []byte, jobID string) (string, error) {
	f, err := os.CreateTemp("", "salesforce-bulk2-"+jobID+"-*.csv")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		_ = os.Remove(f.Name())
		return "", fmt.Errorf("writing Bulk API 2.0 results to %s: %w", f.Name(), err)
	}
	return f.Name(), nil
}
