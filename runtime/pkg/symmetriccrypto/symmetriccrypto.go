package symmetriccrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func GenerateKey(n int) ([]byte, error) {
	res := make([]byte, n)
	_, err := rand.Read(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type Encoder struct {
	block cipher.Block
	gcm   cipher.AEAD
}

func NewEphemeralEncoder(keySize int) (Encoder, error) {
	key, err := GenerateKey(keySize)
	if err != nil {
		return Encoder{}, err
	}

	return NewEncoder(key)
}

func NewEncoder(key []byte) (Encoder, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return Encoder{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Encoder{}, err
	}

	return Encoder{
		block: block,
		gcm:   gcm,
	}, nil
}

func (e Encoder) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	res := e.gcm.Seal(nonce, nonce, data, nil)
	return res, nil
}

func (e Encoder) Decrypt(data []byte) ([]byte, error) {
	nonceSize := e.gcm.NonceSize()
	if len(data) <= nonceSize {
		return nil, errors.New("ciphertext is shorter than the nonce size")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	res, err := e.gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}
