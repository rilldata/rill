package test_utils

import (
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/server"
)

func CreateTestInstance(server *server.Server, dsn string, files []string) (string, error) {
	instanceResp, err := server.CreateInstance(nil, &api.CreateInstanceRequest{
		Driver: "duckdb",
		Dsn:    dsn,
	})
	if err != nil {
		return "", err
	}

	// TODO
	//for _, file := range files {
	//}

	return instanceResp.InstanceId, nil
}
