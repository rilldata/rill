package examples

import (
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Example struct {
	Name        string   `yaml:"-"`
	Description string   `yaml:"description"`
	YAML        string   `yaml:"yaml"`
	Tags        []string `yaml:"_"`
}

var examples = []Example{
	{
		Description: "A model that ingests from a public HTTP URL in duckdb database",
		Name:        "simple_http",
		Tags:        []string{"duckdb", "http"},
		YAML: `
type: model
connector: "https"
uri: "<URI>"`,
	},
	{
		Description: "A model that ingests from parquet files stored in GCS bucket using a glob pattern in duckdb database",
		Name:        "simple_gcs_parquet",
		Tags:        []string{"duckdb", "gcs", "parquet", "gs"},
		YAML: `
type: model
connector: "duckdb"
sql: |
	SELECT * FROM read_parquet("gs://<BUCKET_NAME>/<PATH>/*.parquet")
`,
	},
	{
		Description: "A model that ingests from CSV files in GCS bucket using a glob pattern in duckdb database",
		Name:        "simple_gcs_csv",
		Tags:        []string{"duckdb", "gcs", "csv", "gs"},
		YAML: `
type: model
connector: "duckdb"
sql: |
	SELECT * FROM read_csv("gs://<BUCKET_NAME>/<PATH>/*.csv")
`,
	},
	{
		Description: "A model that ingests from CSV files in GCS bucket using a glob pattern in duckdb database. It also unifies the schema of the CSV files.",
		Name:        "simple_gcs_csv_unify",
		Tags:        []string{"duckdb", "gcs", "csv", "union", "gs"},
		YAML: `
type: model
connector: "duckdb"
sql: |
  SELECT * FROM read_csv("gs://<BUCKET_NAME>/<PATH>/*.csv", union_by_name=true)
`,
	},
}

// scoredExample holds a pointer to an Example and its corresponding match score.
type scoredExample struct {
	Ex    *Example
	Score int
}

// Top3Fuzzy scores each Example by fuzzy-matching its tags against the query
// and returns the top 3 highest-scoring Examples.
func Top3Fuzzy(query string) []Example {
	var scored []scoredExample

	// Compute a simple score: count of tags that fuzzy-match the query
	for i := range examples {
		score := 0
		for _, tag := range examples[i].Tags {
			if fuzzy.RankMatchFold(tag, query) >= 0 {
				score++
			}
		}
		if score > 0 {
			scored = append(scored, scoredExample{Ex: &examples[i], Score: score})
		}
	}

	// Sort by descending score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Select top 3 (or fewer if less available)
	n := 3
	if len(scored) < n {
		n = len(scored)
	}
	result := make([]Example, n)
	for i := 0; i < n; i++ {
		result[i] = *scored[i].Ex
	}
	return result
}
