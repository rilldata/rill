package server

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rilldata/rill-developer/runtime/database"
	"net/http"
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

func SetupRoutes(db *sql.DB, port string) {
	r := gin.New()

	r.POST("/query", func(c *gin.Context) {
		handleQuery(c, db)
	})

	r.POST("/prepare", func(c *gin.Context) {
		prepareQuery(c, db)
	})

	err := r.Run(":" + port)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func handleQuery(c *gin.Context, db *sql.DB) {
	var queryRequest QueryRequest
	err := c.BindJSON(&queryRequest)
	if err != nil {
		returnError(c, err)
		return
	}

	rows, queryError := db.Query(queryRequest.Query)
	if queryError != nil {
		returnError(c, queryError)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": database.ParseRows(rows),
		})
	}
}

func prepareQuery(c *gin.Context, db *sql.DB) {
	var queryRequest QueryRequest
	err := c.BindJSON(&queryRequest)
	if err != nil {
		returnError(c, err)
		return
	}

	_, queryError := db.Prepare(queryRequest.Query)
	if queryError != nil {
		returnError(c, queryError)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	}
}
