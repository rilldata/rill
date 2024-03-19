package metricssqlparser

import (
	"context"
	"testing"

	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCompiler_Compile(t *testing.T) {
	compiler := New()
	runtime, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := runtime.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	olap, release, err := runtime.OLAP(context.Background(), instanceID, "")
	require.NoError(t, err)
	defer release()

	passTests := map[string]string{
		"select pub, dom from ad_bids_metrics":                                                                                          "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\"",
		"select pub, dom from ad_bids_metrics LIMIT 5":                                                                                  "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" LIMIT 5",
		"select pub, dom from ad_bids_metrics order by pub desc":                                                                        "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" ORDER BY \"publisher\" DESC",
		"select pub, dom from ad_bids_metrics order by pub desc, dom asc":                                                               "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" ORDER BY \"publisher\" DESC, \"domain\" ASC",
		"select pub, dom from ad_bids_metrics order by pub desc, dom asc limit 10":                                                      "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" ORDER BY \"publisher\" DESC, \"domain\" ASC LIMIT 10",
		"select pub, dom from ad_bids_metrics where tld = 'Yahoo'":                                                                      "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo'",
		"select pub, dom from ad_bids_metrics where tld = 'Yahoo' LIMIT 5":                                                              "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' LIMIT 5",
		"select pub, dom from ad_bids_metrics  where tld = 'Yahoo' order by pub desc":                                                   "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' ORDER BY \"publisher\" DESC",
		"select pub, dom from ad_bids_metrics  where tld = 'Yahoo' order by pub desc, dom asc":                                          "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' ORDER BY \"publisher\" DESC, \"domain\" ASC",
		"select pub, dom from ad_bids_metrics  where tld = 'Yahoo'order by pub desc, dom asc limit 10":                                  "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' ORDER BY \"publisher\" DESC, \"domain\" ASC LIMIT 10",
		"select pub, dom,measure_0,measure_1 from ad_bids_metrics":                                                                      "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0, avg(bid_price) AS measure_1 FROM \"ad_bids\" GROUP BY \"publisher\", \"domain\"",
		"select pub, dom,measure_0 from ad_bids_metrics LIMIT 5":                                                                        "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY \"publisher\", \"domain\" LIMIT 5",
		"select pub, dom,measure_0 from ad_bids_metrics order by pub desc":                                                              "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC",
		"select pub, dom,measure_0 from ad_bids_metrics order by pub desc, dom asc":                                                     "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC, \"domain\" ASC",
		"select pub, dom,measure_0 from ad_bids_metrics order by pub desc, dom asc limit 10":                                            "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC, \"domain\" ASC LIMIT 10",
		"select pub, dom,measure_0 from ad_bids_metrics where tld = 'Yahoo'":                                                            "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\"",
		"select pub, dom,measure_0 from ad_bids_metrics where tld = 'Yahoo' LIMIT 5":                                                    "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\" LIMIT 5",
		"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' order by pub desc":                                         "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC",
		"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' order by pub desc, dom asc":                                "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC, \"domain\" ASC",
		"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' order by pub desc, dom asc limit 10":                       "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\" ORDER BY \"publisher\" DESC, \"domain\" ASC LIMIT 10",
		"select pub, dom,measure_0 from ad_bids_metrics  where tld = 'Yahoo' having measure_0 > 10 order by pub desc, dom asc limit 10": "SELECT \"publisher\" AS pub, \"domain\" AS dom, count(*) AS measure_0 FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = 'Yahoo' GROUP BY \"publisher\", \"domain\" HAVING count(*) > 10 ORDER BY \"publisher\" DESC, \"domain\" ASC LIMIT 10",
		"select pub, dom from ad_bids_metrics where tld = '{{.user.domain}}'":                                                           "SELECT \"publisher\" AS pub, \"domain\" AS dom FROM \"ad_bids\" WHERE regexp_extract(domain, '(.*\\.)?(.*\\.com)', 2) = '{{.user.domain}}'",
		"select pub, dom as site from ad_bids_metrics":                                                                                  "SELECT \"publisher\" AS pub, \"domain\" AS site FROM \"ad_bids\"",
	}
	for inSQL, outSQL := range passTests {
		got, _, _, err := compiler.Compile(ctrl, instanceID, inSQL, nil)
		require.NoError(t, err)
		if got != outSQL {
			t.Errorf("Compiler.Compile() input = %v, got = %v, want = %v", inSQL, got, outSQL)
		}
		res, err := olap.Execute(context.Background(), &drivers.Statement{Query: got})
		require.NoError(t, err)
		require.NoError(t, res.Close())
	}
}

func TestCompiler_CompileError(t *testing.T) {
	compiler := New()
	runtime, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := runtime.Controller(context.Background(), instanceID)
	require.NoError(t, err)

	sqlToErrMsg := map[string]string{
		"select max(pub), dom from ad_bids_metrics":                         "metrics sql: can only select plain dimension/measures",
		"select pub, dom from ad_bids_metrics where toUpper(pub) = 'Yahoo'": "metrics sql: unsupported expression \"TOUPPER(`pub`)\"",
	}
	for inSQL, errMsg := range sqlToErrMsg {
		_, _, _, err := compiler.Compile(ctrl, instanceID, inSQL, nil)
		require.Error(t, err)
		require.ErrorContains(t, err, errMsg)
	}
}
