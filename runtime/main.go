package main

import (
	"database/sql"
	"fmt"
	_ "github.com/marcboeker/go-duckdb"
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

	server.SetupRoutes(db, os.Args[2])
}
