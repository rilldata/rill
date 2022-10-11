package misc

// ConnectorOpenOptions abstracts common fields during an open connection to connector
type ConnectorOpenOptions interface {
	GetRemotePath() string
	// OptionsDummy is there to avoid satisfying other interfaces
	OptionsDummy()
}

type EmptyConnectorOpenOptions struct{}

func (o *EmptyConnectorOpenOptions) GetRemotePath() string {
	return ""
}

func (o *EmptyConnectorOpenOptions) OptionsDummy() {}

// ConnectorIngestOptions abstracts common fields during an ingestion from connector
type ConnectorIngestOptions interface {
	GetSourceName() string
	GetPath() string
}

type BasicConnectorIngestOptions struct {
	SourceName string
	Path       string
}

func (o *BasicConnectorIngestOptions) GetSourceName() string {
	return o.SourceName
}

func (o *BasicConnectorIngestOptions) GetPath() string {
	return o.Path
}
