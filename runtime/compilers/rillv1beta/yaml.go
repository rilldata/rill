package rillv1beta

type Source struct {
	Type         string
	URI          string `yaml:"uri,omitempty"`
	Path         string `yaml:"path,omitempty"`
	Region       string `yaml:"region,omitempty"`
	CSVDelimiter string `yaml:"csv.delimiter,omitempty"`
}
