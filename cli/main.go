package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/azureblob"
)

const (
	// The storage container to access.
	containerName = "rill-test"
)

func main() {
	name := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_KEY")
	credential, err := azblob.NewSharedKeyCredential(name, key)
	if err != nil {
		log.Fatal(err)
	}

	containerURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", name, containerName)
	client, err := container.NewClientWithSharedKeyCredential(containerURL, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	bkt, err := azureblob.OpenBucket(context.Background(), client, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer bkt.Close()

	// Now we can use b to read or write files to the container.
	data, err := bkt.ReadAll(context.Background(), "AdBids.parquet")
	if err != nil {
		panic(err)
	}
	_ = data

}
