package main

import (
	"database/sql"
	"fmt"
	_ "github.com/rilldata/rill-developer/runtime/duckdb"
	"github.com/rilldata/rill-developer/runtime/server"
	"os"
)

func main() {
	duckDbFile := os.Args[1] + "?access_mode=READ_WRITE"
	db, err := sql.Open("duckdb", duckDbFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Started duckdb server. DB File=%s\n", duckDbFile)

	server.SetupRoutes(db, os.Args[2])
}
