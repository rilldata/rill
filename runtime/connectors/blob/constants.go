package blob

type BlobType string

const (
	File BlobType = "file"
	S3   BlobType = "s3"
	GCS  BlobType = "gcs"
)
