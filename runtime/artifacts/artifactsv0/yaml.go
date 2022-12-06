package artifactsv0

type Source struct {
	Version      string
	Type         string
	URI          string `yaml:"uri,omitempty"`
	Path         string `yaml:"path,omitempty"`
	Region       string `yaml:"region,omitempty"`
	CSVDelimiter string `yaml:"csv.delimiter,omitempty"`
}
