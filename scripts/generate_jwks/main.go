package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"gopkg.in/square/go-jose.v2"
)

func main() {
	// NOTE: JWKS generation based on: https://github.com/go-jose/go-jose/blob/v3/jose-util/generate.go

	// Generate RSA private key
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Create JWK
	jwk := jose.JSONWebKey{
		Key:       rsaKey,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	// Set key ID based on JWK thumbprint
	thumb, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		log.Fatal(err.Error())
	}
	jwk.KeyID = base64.URLEncoding.EncodeToString(thumb)

	// Create JWKS JSON
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Print
	fmt.Printf("RILL_ADMIN_SIGNING_KEY_ID=%s\n", jwk.KeyID)
	fmt.Printf("RILL_ADMIN_SIGNING_JWKS=%s\n", string(jwksJSON))
}
