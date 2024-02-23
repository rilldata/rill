package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type motherduckToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.Handle
	logger *zap.Logger
}

var _ drivers.Transporter = &motherduckToDuckDB{}

func NewMotherduckToDuckDB(from drivers.Handle, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &motherduckToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

// TODO: should it run count from user_query to set target in progress ?
func (t *motherduckToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	srcCfg, err := parseDBSourceProperties(srcProps)
	if err != nil {
		return err
	}

	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	// we first ingest data in a temporary table in the main db
	// and then copy it to the final table to ensure that the final table is always created using CRUD APIs which takes care
	// whether table goes in main db or in separate table specific db
	tmpTable := fmt.Sprintf("__%s_tmp_motherduck", sinkCfg.Table)
	defer func() {
		// ensure temporary table is cleaned
		err := t.to.Exec(context.Background(), &drivers.Statement{
			Query:       fmt.Sprintf("DROP TABLE IF EXISTS %s", tmpTable),
			Priority:    100,
			LongRunning: true,
		})
		if err != nil {
			t.logger.Error("failed to drop temp table", zap.String("table", tmpTable), zap.Error(err))
		}
	}()

	config := t.from.Config()
	err = t.to.WithConnection(ctx, 1, true, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
		res, err := t.to.Execute(ctx, &drivers.Statement{Query: "SELECT current_database(),current_schema();"})
		if err != nil {
			return err
		}

		var localDB, localSchema string
		for res.Next() {
			if err := res.Scan(&localDB, &localSchema); err != nil {
				_ = res.Close()
				return err
			}
		}
		_ = res.Close()

		// get token
		token, _ := config["token"].(string)
		if token == "" && config["allow_host_access"].(bool) {
			token = os.Getenv("motherduck_token")
		}
		if token == "" {
			return fmt.Errorf("no motherduck token found. Refer to this documentation for instructions: https://docs.rilldata.com/deploy/credentials/motherduck")
		}

		// load motherduck extension; connect to motherduck service
		err = t.to.Exec(ctx, &drivers.Statement{Query: "INSTALL 'motherduck'; LOAD 'motherduck';"})
		if err != nil {
			return fmt.Errorf("failed to load motherduck extension %w", err)
		}

		if err = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("SET motherduck_token='%s'", token)}); err != nil {
			if !strings.Contains(err.Error(), "can only be set during initialization") {
				return fmt.Errorf("failed to set motherduck token %w", err)
			}
		}

		if err = t.to.Exec(ctx, &drivers.Statement{Query: "ATTACH 'md:'"}); err != nil {
			if !strings.Contains(err.Error(), "already attached") {
				return fmt.Errorf("failed to connect to motherduck %w", err)
			}
		}

		var names []string

		db := srcCfg.Database
		if db == "" {
			// get list of all motherduck databases
			res, err = t.to.Execute(ctx, &drivers.Statement{Query: "SELECT name FROM md_databases();"})
			if err != nil {
				return err
			}
			defer res.Close()

			for res.Next() {
				var name string
				if res.Scan(&name) != nil {
					return err
				}
				names = append(names, name)
			}
			// single motherduck db, use db to allow user to run query without specifying db name
			if len(names) == 1 {
				db = names[0]
			}
		}

		if db != "" {
			err = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("USE %s;", safeName(db))})
			if err != nil {
				return err
			}

			defer func(ctx context.Context) { // revert back to localdb
				err = t.to.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("USE %s.%s;", safeName(localDB), safeName(localSchema))})
				if err != nil {
					t.logger.Error("failed to switch to local database", zap.Error(err))
				}
			}(ensuredCtx)
		}

		if srcCfg.SQL == "" {
			return fmt.Errorf("property \"sql\" is mandatory for connector \"motherduck\"")
		}

		userQuery := strings.TrimSpace(srcCfg.SQL)
		userQuery, _ = strings.CutSuffix(userQuery, ";") // trim trailing semi colon
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s.%s.%s AS (%s\n);", safeName(localDB), safeName(localSchema), safeName(tmpTable), userQuery)
		return t.to.Exec(ctx, &drivers.Statement{Query: query})
	})
	if err != nil {
		return err
	}

	// copy data from temp table to target table
	return t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", tmpTable))
}
