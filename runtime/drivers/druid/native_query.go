package druid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

var druidSQLDSN = regexp.MustCompile(`/v2/sql/?`)

type NativeClient struct {
	client     *http.Client
	dsn        string
	logger     *zap.Logger
	logQueries bool
}

func NewNativeClient(olap drivers.OLAPStore) (*NativeClient, error) {
	conn, ok := olap.(*connection)
	if !ok {
		return nil, fmt.Errorf("invalid handle type, not a druid connection")
	}

	dsn, err := dsnFromConfig(conn.config)
	if err != nil {
		return nil, err
	}

	if dsn == "" {
		return nil, fmt.Errorf("druid connector config not found in instance")
	}
	dsn = druidSQLDSN.ReplaceAllString(dsn, "/v2/")
	return &NativeClient{
		client:     &http.Client{},
		dsn:        dsn,
		logger:     conn.logger,
		logQueries: conn.config.LogQueries,
	}, nil
}

func (n *NativeClient) Search(ctx context.Context, dr *NativeSearchQueryRequest) (NativeSearchQueryResponse, error) {
	logger := n.logger.With(zap.String("query_id", dr.Context.QueryID))
	if n.logQueries {
		logger.Info("Executing native query", zap.Any("request", dr))
	}
	b, err := json.Marshal(dr)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(b)

	context.AfterFunc(ctx, func() {
		tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r, err := http.NewRequestWithContext(tctx, http.MethodDelete, n.dsn+"/"+dr.Context.QueryID, http.NoBody)
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
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("druid native query failed with status code: %d", resp.StatusCode)
	}
	var nativeResponse NativeSearchQueryResponse
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&nativeResponse)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	if n.logQueries {
		logger.Info("Druid native query successful", zap.Any("response", nativeResponse))
	}
	return nativeResponse, nil
}

type NativeSearchQueryRequest struct {
	Context          QueryContext           `json:"context"`
	QueryType        string                 `json:"queryType"`
	DataSource       string                 `json:"dataSource"`
	SearchDimensions []string               `json:"searchDimensions"`
	VirtualColumns   []NativeVirtualColumns `json:"virtualColumns"`
	Limit            int                    `json:"limit"`
	Query            NativeSearchQuery      `json:"query"`
	Sort             NativeSearchSort       `json:"sort"`
	Intervals        []string               `json:"intervals"`
	Filter           map[string]interface{} `json:"filter"`
}

type QueryContext struct {
	QueryID string `json:"queryId,omitempty"`
}

type NativeVirtualColumns struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

type NativeSearchQuery struct {
	Type          string `json:"type"`
	CaseSensitive bool   `json:"case_sensitive"`
	Value         string `json:"value"`
}

type NativeSearchSort struct {
	Type string `json:"type"`
}

type NativeSearchQueryResponse []struct {
	Timestamp time.Time `json:"timestamp"`
	Result    []struct {
		Dimension string `json:"dimension"`
		Value     string `json:"value"`
	} `json:"result"`
}

type QueryPlan struct {
	Query struct {
		Filter *map[string]interface{} `json:"filter"`
	} `json:"query"`
}

func NewNativeSearchQueryRequest(source, search string, dims []string, virtualCols []NativeVirtualColumns, limit int, start, end time.Time, filter map[string]interface{}) NativeSearchQueryRequest {
	return NativeSearchQueryRequest{
		Context: QueryContext{
			QueryID: uuid.New().String(),
		},
		QueryType:        "search",
		DataSource:       source,
		SearchDimensions: dims,
		VirtualColumns:   virtualCols,
		Limit:            limit,
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
