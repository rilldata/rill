package clickhouse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

var dbPath = "/Users/kanshul/automatadmore_data/user_files"

type transporter struct {
	from   drivers.FileStore
	to     drivers.OLAPStore
	logger *zap.Logger
}

var _ drivers.Transporter = &transporter{}

func NewFileStoreToClickHouse(store drivers.FileStore, olap drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &transporter{
		from:   store,
		to:     olap,
		logger: logger,
	}
}

// Transfer implements drivers.Transporter.
func (t *transporter) Transfer(ctx context.Context, srcProps map[string]any, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	srcCfg, err := parseFileSourceProperties(srcProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	localPaths, err := t.from.FilePaths(ctx, srcProps)
	if err != nil {
		return err
	}

	if len(localPaths) == 0 {
		return fmt.Errorf("no files to ingest")
	}

	// size := fileSize(localPaths)
	// if !sizeWithinStorageLimits(t.to, size) {
	// 	return drivers.ErrStorageLimitExceeded
	// }
	// opts.Progress.Target(size, drivers.ProgressUnitByte)

	pathDir := filepath.Dir(localPaths[0])
	symLinkDirName := fmt.Sprintf("link_%v", time.Now().Unix())
	symlinkPath := filepath.Join(dbPath, symLinkDirName)
	if err := os.Symlink(pathDir, symlinkPath); err != nil {
		return err
	}
	defer func() {
		err := os.Remove(symlinkPath)
		if err != nil {
			t.logger.Error("failed to remove symlink", zap.Error(err))
		}
	}()

	var formatString string
	if srcCfg.Format != "" {
		formatString = fmt.Sprintf(", %s", srcCfg.Format)
	}
	err = t.to.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf("CREATE OR REPLACE TABLE %s ENGINE = MergeTree ORDER BY tuple() AS SELECT * FROM file('%s' %s)",
			safeSQLName(sinkCfg.Table),
			filepath.Join(symLinkDirName, filepath.Base(localPaths[0])),
			formatString),
	})
	if err != nil {
		return err
	}

	for i := 1; i < len(localPaths); i++ {
		err = t.to.Exec(ctx, &drivers.Statement{
			Query: fmt.Sprintf("INSERT INTO %s SELECT * FROM file('%s' %s)", safeSQLName(sinkCfg.Table), filepath.Join(symLinkDirName, filepath.Base(localPaths[i])), formatString),
		})
		if err != nil {
			return err
		}
	}
	return nil

}
