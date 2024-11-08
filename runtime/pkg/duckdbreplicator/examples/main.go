package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	duckdbreplicator "github.com/rilldata/duckdb-replicator"
	_ "gocloud.dev/blob/gcsblob"
)

func main() {
	backup, err := duckdbreplicator.NewGCSBackupProvider(context.Background(), &duckdbreplicator.GCSBackupProviderOptions{
		UseHostCredentials: true,
		Bucket:             "<my_bucket>",
		UniqueIdentifier:   "756c6367-e807-43ff-8b07-df1bae29c57e/",
	})
	if err != nil {
		panic(err)
	}

	dbOptions := &duckdbreplicator.DBOptions{
		LocalPath:      "<local-path>",
		BackupProvider: backup,
		ReadSettings:   map[string]string{"memory_limit": "2GB", "threads": "1"},
		WriteSettings:  map[string]string{"memory_limit": "8GB", "threads": "2"},
		InitQueries:    []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         slog.Default(),
	}

	db, err := duckdbreplicator.NewDB(context.Background(), "756c6367-e807-43ff-8b07-df1bae29c57e", dbOptions)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	t := time.Now()
	// create table
	err = db.CreateTableAsSelect(context.Background(), "test-2", `SELECT * FROM read_parquet('data*.parquet')`, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("time taken %v\n", time.Since(t))

	// rename table
	err = db.RenameTable(context.Background(), "test-2", "test")
	if err != nil {
		panic(err)
	}

	// insert into renamed table
	err = db.InsertTableAsSelect(context.Background(), "test", `SELECT * FROM read_parquet('data*.parquet')`, &duckdbreplicator.InsertTableOptions{
		Strategy: duckdbreplicator.IncrementalStrategyAppend,
	})
	if err != nil {
		panic(err)
	}

	// get count
	conn, release, err := db.AcquireReadConnection(context.Background())
	if err != nil {
		panic(err)
	}
	defer release()

	var count int
	err = conn.Connx().QueryRowxContext(context.Background(), `SELECT count(*) FROM "test"`).Scan(&count)
	if err != nil {
		fmt.Printf("error %v\n", err)
	}
	fmt.Println(count)

}
