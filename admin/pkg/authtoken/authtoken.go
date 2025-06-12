package authtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// ErrMalformed is used when attempting to parse an invalid token string.
var ErrMalformed = errors.New("malformed auth token")

// Prefix is prepended to every auth token.
const Prefix = "rill"

// Type is part of the token prefix in the string representation.
type Type string

const (
	TypeUser       Type = "usr"
	TypeService    Type = "svc"
	TypeDeployment Type = "dpl"
	TypeMagic      Type = "mgc"
)

// Validate checks that the type is a known enum value.
func (t Type) Validate() bool {
	switch t {
	case TypeUser, TypeService, TypeDeployment, TypeMagic:
		return true
	default:
		return false
	}
}

// Token is a parsed authentication token with a type, UUID ID, and 24-byte secret.
// Tokens can be (de)serialized as strings.
// Example string representation of a user token: rill_usr_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK.
type Token struct {
	Type   Type
	ID     uuid.UUID
	Secret [24]byte
}

// NewRandom generates a securely random token.
func NewRandom(t Type) *Token {
	tkn := Token{
		Type: t,
		ID:   uuid.New(),
	}

	if _, err := rand.Read(tkn.Secret[:]); err != nil {
		panic(err)
	}

	return &tkn
}

// FromString re-creates a token from it's string representation (acquired by calling String()).
// The things I do for pretty tokens.
func FromString(s string) (*Token, error) {
	parts := strings.Split(s, "_")
	if len(parts) != 3 {
		return nil, ErrMalformed
	}

	if parts[0] != Prefix {
		return nil, ErrMalformed
	}

	typ := Type(parts[1])
	if !typ.Validate() {
		return nil, ErrMalformed
	}

	payload, ok := unmarshalBase62(parts[2])
	if !ok {
		return nil, ErrMalformed
	}

	if len(payload) > 40 {
		return nil, ErrMalformed
	} else if len(payload) < 40 {
		payload = padLeft(payload, 40)
	}

	var id [16]byte
	copy(id[:], payload[0:16])

	var secret [24]byte
	copy(secret[:], payload[16:40])

	return &Token{
		Type:   typ,
		ID:     uuid.UUID(id),
		Secret: secret,
	}, nil
}

func FromParts(typ Type, id uuid.UUID, secret []byte) (*Token, error) {
	if !typ.Validate() {
		return nil, errors.New("invalid token type")
	}

	if len(secret) != 24 {
		return nil, errors.New("secret must be 24 bytes")
	}

	return &Token{
		Type:   typ,
		ID:     id,
		Secret: [24]byte(secret),
	}, nil
}

// FromID returns a Token with the given type and id, and a zero secret.
func FromID(typ Type, id uuid.UUID) *Token {
	return &Token{
		Type:   typ,
		ID:     id,
		Secret: [24]byte{},
	}
}

// String canonically encodes the token as string.
func (t *Token) String() string {
	payload := make([]byte, 40)
	copy(payload[0:16], t.ID[:])
	copy(payload[16:40], t.Secret[:])
	return fmt.Sprintf("%s_%s_%s", Prefix, t.Type, marshalBase62(payload))
}

// SecretHash returns a SHA256 hash of the token secret.
func (t *Token) SecretHash() []byte {
	hashed := sha256.Sum256(t.Secret[:])
	return hashed[:]
}

// Prefix returns a safe, partial display string for the token, e.g., "rill_usr_abcdefghij".
// This works even if t.Secret is empty (all zeroes), as long as t.ID is set.
func (t *Token) Prefix() string {
	payload := make([]byte, 40)
	copy(payload[0:16], t.ID[:])

	n := len(Prefix) + len(t.Type) + 12
	return t.String()[0:n]
}

// MatchByPrefix attempts to match a decoded token by its prefix.
func MatchByPrefix(input string, tokens []*Token) []*Token {
	if len(input) < 10 || !strings.HasPrefix(input, Prefix+"_") {
		return nil
	}
	var matches []*Token
	for _, t := range tokens {
		if strings.HasPrefix(t.Prefix(), input) {
			matches = append(matches, t)
		}
	}
	return matches
}

// marshalBase62 marshals a byte slice to a string of [0-9A-Za-z] characters.
func marshalBase62(val []byte) string {
	var i big.Int
	i.SetBytes(val)
	return i.Text(62)
}

// unmarshalBase62 unmarshals a byte slice encoded with marshalBase62.
func unmarshalBase62(s string) ([]byte, bool) {
	var i big.Int
	_, ok := i.SetString(s, 62)
	if !ok {
		return nil, false
	}
	return i.Bytes(), true
}

// padLeft pads a byte slice with zeros on the left such that its length is n.
// If the slice is already longer than n, it is returned as is.
func padLeft(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}

	padded := make([]byte, n)
	copy(padded[n-len(b):], b)
	return padded
}
