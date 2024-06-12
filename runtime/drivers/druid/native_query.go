package druid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type QueryContext struct {
	QueryID string `json:"queryId,omitempty"`
}

type NativeSearchQuery struct {
	Type          string `json:"type"`
	CaseSensitive bool   `json:"case_sensitive"`
	Value         string `json:"value"`
}
type NativeSearchSort struct {
	Type string `json:"type"`
}
type NativeSearchQueryRequest struct {
	Context          QueryContext           `json:"context"`
	QueryType        string                 `json:"queryType"`
	DataSource       string                 `json:"dataSource"`
	SearchDimensions []string               `json:"searchDimensions"`
	Limit            int                    `json:"limit"`
	Query            NativeSearchQuery      `json:"query"`
	Sort             NativeSearchSort       `json:"sort"`
	Intervals        []string               `json:"intervals"`
	Filter           map[string]interface{} `json:"filter"`
}

type NativeSearchQueryResponse []struct {
	Timestamp time.Time `json:"timestamp"`
	Result    []struct {
		Dimension string `json:"dimension"`
		Value     string `json:"value"`
	} `json:"result"`
}

type NativeQuery struct {
	client *http.Client
	dsn    string
}

func NewNativeQuery(dsn string) NativeQuery {
	return NativeQuery{
		client: &http.Client{},
		dsn:    dsn,
	}
}

type QueryPlan struct {
	Query struct {
		Filter *map[string]interface{} `json:"filter"`
	} `json:"query"`
}

func (n *NativeQuery) Do(ctx context.Context, dr, res interface{}, queryID string) error {
	b, err := json.Marshal(dr)
	if err != nil {
		return err
	}

	bodyReader := bytes.NewReader(b)

	context.AfterFunc(ctx, func() {
		tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r, err := http.NewRequestWithContext(tctx, http.MethodDelete, n.dsn+"/"+queryID, http.NoBody)
		if err != nil {
			return
		}

		resp, err := n.client.Do(r)
		if err != nil {
			return
		}
		resp.Body.Close()
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.dsn, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&res)
	if err != nil {
		resp.Body.Close()
		return err
	}
	return nil
}

func NewNativeSearchQueryRequest(source, search string, dimensions []string, start, end time.Time, filter map[string]interface{}) NativeSearchQueryRequest {
	return NativeSearchQueryRequest{
		Context: QueryContext{
			QueryID: uuid.New().String(),
		},
		QueryType:        "search",
		DataSource:       source,
		SearchDimensions: dimensions,
		Limit:            100,
		Query: NativeSearchQuery{
			Type:          "contains",
			CaseSensitive: false,
			Value:         search,
		},
		Sort: NativeSearchSort{
			Type: "lexicographic",
		},
		Intervals: []string{
			fmt.Sprintf("%s/%s", start.Format(time.RFC3339), end.Format(time.RFC3339)),
		},
		Filter: filter,
	}
}
