package resolvers

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func Test_parsedSQL(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)

	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			"simple",
			"select * from ad_bids_metrics",
			"select * FROM \"ad_bids\"",
		},
		{
			"simple quoted",
			"select * from \"ad_bids_metrics\"",
			"select * FROM \"ad_bids\"",
		},
		{
			"aggregate",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(num_bids),AGGREGATE(avg_bid_price) FROM ad_bids_metrics GROUP BY ALL",
			"SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
		},
		{
			"aggregate with mv appended",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(ad_bids_metrics.num_bids),AGGREGATE(ad_bids_metrics.avg_bid_price) FROM ad_bids_metrics GROUP BY ALL",
			"SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
		},
		{
			"aggregate with mv appended and quoted",
			"SELECT pub,domain_parts,dom,tld,null_publisher,AGGREGATE(\"ad_bids_metrics\".\"num_bids\"),AGGREGATE(ad_bids_metrics.\"avg_bid_price\") FROM ad_bids_metrics GROUP BY ALL",
			"SELECT pub,domain_parts,dom,tld,null_publisher,count(*),avg(bid_price) FROM \"ad_bids\" GROUP BY ALL",
		},
		{
			"aggregate and spaces",
			`SELECT pub,dom,AGGREGATE("bid's number"),AGGREGATE("total volume"),Aggregate("total click""s") From ad_bids_mini_metrics GROUP BY ALL`,
			"SELECT pub,dom,count(*),sum(volume),sum(clicks) FROM \"ad_bids_mini\" GROUP BY ALL",
		},
		{
			"aggregate and join",
			`with a as (
				select
					publisher,
					AGGREGATE(ad_bids_mini_metrics."total volume") as total_volume,
					AGGREGATE(ad_bids_mini_metrics."total click""s") as total_clicks
				from
					ad_bids_mini_metrics
				group by
					publisher
				),
				b as (
				select
					publisher,
					AGGREGATE(ad_bids_metrics."avg_bid_price") as avg_bids
				from
					ad_bids_metrics
				group by
					publisher
				)
				select
					a.publisher,
					a.total_volume,
					a.total_clicks,
					b.avg_bids
				from
					a
				join b on
					a.publisher = b.publisher
				`,
			`with a as (
					select
						publisher,
						sum(volume) as total_volume,
						sum(clicks) as total_clicks
					FROM "ad_bids_mini"
					group by
						publisher
					),
					b as (
					select
						publisher,
						avg(bid_price) as avg_bids
					FROM "ad_bids"
					group by
						publisher
					)
					select
						a.publisher,
						a.total_volume,
						a.total_clicks,
						b.avg_bids
					from
						a
					join b on
						a.publisher = b.publisher
					`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsedSQL(context.Background(), ctrl, tt.sql)
			require.NoError(t, err)
			got = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(got, "\n", " "), "\t", " "), " ")
			tt.want = regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(tt.want, "\n", " "), "\t", " "), " ")
			if got != tt.want {
				t.Errorf("parsedSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
