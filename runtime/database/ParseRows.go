package database

import (
	"database/sql"
	"fmt"
	"strconv"
)

func ParseRows(rows *sql.Rows) []map[string]interface{} {
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
	err = rows.Err()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer rows.Close()
	return results
}
