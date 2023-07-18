package emitter

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/zap"
)

// GCSSink sinks events to a GCS bucket.
// Events are partitioned by organization_id and project_id:
// 1. If both are defined then the full path is bucketName/organization_id/project_id/fileName
// 2. If only organization_id is defined then the full path is bucketName/organization_id/fileName
// 3. Otherwise, the full path is bucketName/fileName
// fileName is unique on each sink call
type GCSSink struct {
	bucket string
	client *storage.Client
	logger *zap.Logger
}

func NewGCSSink(bucketName string, logger *zap.Logger) (*GCSSink, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}
	return &GCSSink{bucket: bucketName, client: client, logger: logger}, nil
}

func (s *GCSSink) Sink(events []Event) {
	// Create a map to hold group of temp file paths
	tmpFiles := make(map[string]*os.File)

	// iterate over events
	for _, event := range events {
		orgID := ""
		projectID := ""

		// Find the organization and project id in the dims
		for _, dim := range event.Dims {
			if dim.Name == "organization_id" {
				orgID = dim.Value
			} else if dim.Name == "project_id" {
				projectID = dim.Value
			}
		}

		// Create a composite key based on the org and project id
		key := orgID
		if projectID != "" {
			key = key + ";" + projectID
		}

		// Check if the file is already created for the key, if not create one
		tmpFile, ok := tmpFiles[key]
		if !ok {
			tmpFile, err := os.CreateTemp("", "events")
			if err != nil {
				s.logger.Debug(fmt.Sprintf("could not create a temp file for event: %v", event), zap.Error(err))
				return
			}
			tmpFiles[key] = tmpFile
		}

		// Write event to the temp file
		bytes, err := convertEventToBytes(event)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not serialize event: %v", event), zap.Error(err))
			continue
		}
		_, err = tmpFile.Write(bytes)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not append event to a temp file: %v", event), zap.Error(err))
			continue
		}
		_, err = tmpFile.WriteString("\n")
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not append a separator to a temp file after event: %v", event), zap.Error(err))
		}
	}

	// Iterate over the tmpFiles map and upload them
	for key, tmpFile := range tmpFiles {
		// Separate the orgID and projectID
		ids := strings.Split(key, ";")
		orgID := ids[0]
		projectID := ""
		if len(ids) > 1 {
			projectID = ids[1]
		}

		// Upload the file
		err := s.uploadFile(context.Background(), tmpFile, orgID, projectID)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not upload a file with usage events with a key: %s", key), zap.Error(err))
		}
	}
}

func (s *GCSSink) uploadFile(ctx context.Context, f *os.File, orgID, projectID string) error {
	defer os.Remove(f.Name())

	randomSfx, err := randomHex(2)
	if err != nil {
		return err
	}

	fileName := time.Now().Format(time.RFC3339) + randomSfx + ".json"

	if orgID != "" && projectID != "" {
		fileName = orgID + "/" + projectID + "/" + fileName
	} else if orgID != "" {
		fileName = orgID + "/" + fileName
	}

	wc := s.client.Bucket(s.bucket).Object(fileName).NewWriter(ctx)
	if _, err := io.Copy(wc, f); err != nil {
		return err
	}

	err = wc.Close()
	return err
}

func (s *GCSSink) Close() error {
	return s.client.Close()
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
