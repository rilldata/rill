package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/marcboeker/go-duckdb"
	"net/http"
	"os"
	"strconv"
)

type QueryRequest struct {
	Query string `json:"query"`
}

func returnError(c *gin.Context, err error) {
	fmt.Println(err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": fmt.Sprintf("%s", err),
	})
}

func parseRows(rows *sql.Rows) []map[string]interface{} {
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	results := make([]map[string]interface{}, 0)

	for rows.Next() {
		result := make(map[string]interface{})

		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		for i, value := range values {
			switch value.(type) {
			case nil:
				result[columns[i]] = nil

			case []byte:
				s := string(value.([]byte))
				x, err := strconv.Atoi(s)

				if err != nil {
					result[columns[i]] = s
				} else {
					result[columns[i]] = x
				}

			default:
				result[columns[i]] = value
			}
		}
		results = append(results, result)
	}
	defer rows.Close()
	return results
}

func main() {
	duckDbFile := os.Args[1] + "?access_mode=READ_WRITE"
	db, err := sql.Open("duckdb", duckDbFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	r := gin.Default()
	r.POST("/query", func(c *gin.Context) {
		var queryRequest QueryRequest
		err := c.BindJSON(&queryRequest)
		if err != nil {
			returnError(c, err)
			return
		}
		fmt.Printf("Recieved query: %s\n", queryRequest.Query)
		rows, queryError := db.Query(queryRequest.Query)
		if queryError != nil {
			returnError(c, queryError)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"data": parseRows(rows),
			})
		}
	})
	err = r.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
}
