package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	// Generate session key pairs for https://pkg.go.dev/github.com/gorilla/sessions#NewCookieStore

	// Generate authentication key (64 bytes)
	authKeyBytes := make([]byte, 64)
	if _, err := rand.Read(authKeyBytes); err != nil {
		log.Fatal(err.Error())
	}
	authKey := hex.EncodeToString(authKeyBytes)

	// Generate encryption key (32 bytes)
	encryptionKeyBytes := make([]byte, 32)
	if _, err := rand.Read(encryptionKeyBytes); err != nil {
		log.Fatal(err.Error())
	}
	encryptionKey := hex.EncodeToString(encryptionKeyBytes)

	// Print
	fmt.Printf("RILL_ADMIN_SESSION_KEY_PAIRS=%s,%s\n", authKey, encryptionKey)
}
