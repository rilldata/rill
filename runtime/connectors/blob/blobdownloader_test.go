package blob

// import (
// 	"context"
// 	"fmt"
// 	"reflect"
// 	"testing"

// 	"gocloud.dev/blob"
// 	_ "gocloud.dev/blob/memblob"
// )

// func getBucket() *blob.Bucket {
// 	ctx := context.Background()
// 	bucket, err := blob.OpenBucket(ctx, "mem://")
// 	if err == nil {
// 		// Open the key "foo.txt" for writing with the default options.
// 		for _, year := range []int{2020, 2021} {
// 			for _, month := range []int{1, 2, 3} {
// 				for _, day := range []int{1, 2, 3} {
// 					filename := fmt.Sprintf("%v/%v/%v/%v_%v_%v.csv", year, month, day, year, month, day)
// 					w, _ := bucket.NewWriter(ctx, filename, nil)
// 					w.Close()
// 				}
// 			}
// 		}
// 	}
// 	return bucket
// }
// func TestFetchBlobHandler(t *testing.T) {
// 	bucket := getBucket()
// 	defer bucket.Close()
// 	type args struct {
// 		ctx         context.Context
// 		bucket      *blob.Bucket
// 		config      FetchConfigs
// 		globPattern string
// 		bucketPath  string
// 	}

// 	arg := args{
// 		ctx:         context.Background(),
// 		bucket:      bucket,
// 		config:      FetchConfigs{MaxSize: 10 * 1024 * 1024, MaxDownload: 100, MaxIterations: 1},
// 		globPattern: "2020/01/01/2020_01_01.csv",
// 		bucketPath:  "mem://",
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *BlobHandler
// 		wantErr bool
// 	}{{
// 		name:    "single_file",
// 		args:    arg,
// 		want:    &BlobHandler{bucket: bucket, FileNames: []string{"2020/01/01/2020_01_01.csv"}, BlobType: File, path: },
// 		wantErr: bool(true),
// 	}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := FetchBlobHandler(tt.args.ctx, tt.args.bucket, tt.args.config, tt.args.globPattern, tt.args.bucketPath)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FetchBlobHandler() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FetchBlobHandler() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
