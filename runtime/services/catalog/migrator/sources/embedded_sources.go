package sources

import (
	"fmt"
	"regexp"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	protocolExtraction = regexp.MustCompile(`^(\w*?)://(.*)$`)
	sanitiser          = regexp.MustCompile(`(?im)[:/?\-.~]`)
)

func GetEmbeddedSource(path string) (*runtimev1.Source, bool) {
	path = strings.TrimSpace(strings.Trim(path, `"'`))
	matches := protocolExtraction.FindStringSubmatch(path)
	var connector string
	if len(matches) < 3 {
		if strings.Contains(path, "/") {
			connector = "local_file"
		} else {
			return nil, false
		}
	} else {
		switch matches[1] {
		case "http", "https":
			connector = "https"
		case "s3":
			connector = "s3"
		case "gs":
			connector = "gcs"
		default:
			return nil, false
		}
	}
	name := fmt.Sprintf("%s_%s", connector, sanitiser.ReplaceAllString(path, "_"))
	props, err := structpb.NewStruct(map[string]any{
		"path": path,
	})
	if err != nil {
		// shouldn't happen
		return nil, false
	}
	return &runtimev1.Source{
		Name:       name,
		Connector:  connector,
		Properties: props,
	}, true
}
