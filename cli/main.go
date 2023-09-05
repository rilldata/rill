package main

import (
	"context"
	"fmt"
	"os"

	"gocloud.dev/blob/azureblob"
)

const (
	// The storage container to access.
	containerName = "rill-test"
)

func main() {
	// Construct the service URL.
	// There are many forms of service URLs, see ServiceURLOptions.
	opts := azureblob.NewDefaultServiceURLOptions()
	serviceURL, err := azureblob.NewServiceURL(opts)
	if err != nil {
		panic(err)
	}

	fmt.Println(os.Getenv("AZURE_STORAGE_SAS_TOKEN"))

	// There are many ways to authenticate to Azure.
	// This approach uses environment variables as described in azureblob package
	// documentation.
	// For example, to use shared key authentication, you would set
	// AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY.
	// To use a SAS token, you would set AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_SAS_TOKEN.
	// You can also construct a client using the azblob constructors directly, like
	// azblob.NewServiceClientWithSharedKey.
	client, err := azureblob.NewDefaultClient(serviceURL, containerName)
	if err != nil {
		 panic(err)
	}

	// Create a *blob.Bucket.
	b, err := azureblob.OpenBucket(context.Background(), client, nil)
	if err != nil {
		 panic(err)
	}
	defer b.Close()

	// Now we can use b to read or write files to the container.
	data, err := b.ReadAll(context.Background(), "AdBids.parquet")
	if err != nil {
		 panic(err)
	}
	_ = data

}
