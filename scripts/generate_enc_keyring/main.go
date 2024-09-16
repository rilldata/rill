package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rilldata/rill/admin/database"
)

// This script generates a new encryption keyring for the database and prints it to stdout. Existing keyring can be provided as an argument like:
// go run scripts/generate_enc_keyring/main.go '[{"key_id":"<>","key":"<>"},{"key_id":"key2","key":"<>"}]'
// New key will be prepended to the existing keyring.
func main() {
	var encKeyring []*database.EncryptionKey
	var err error
	if len(os.Args) > 1 {
		encKeyring, err = database.ParseEncryptionKeyring(os.Args[1])
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	// Generate encryption keyring for db column encryption
	newKeyRing, err := database.NewRandomKeyring()
	if err != nil {
		log.Fatal(err.Error())
	}
	newKeyRing = append(newKeyRing, encKeyring...)

	conf, err := json.Marshal(newKeyRing)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("RILL_ADMIN_DATABASE_ENCRYPTION_KEYRING='%s'\n", conf)
}
