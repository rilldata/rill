package securetoken

import "github.com/gorilla/securecookie"

// Codec provides a means of encoding and decoding values to URL-friendly opaque tokens that clients cannot tamper with.
// It encodes values using encoding/gob before encrypting them, and uses base64 URL encoding for serializing the encrypted bytes to the token string.
// It's implemented as a thin wrapper around gorilla/securecookie, which already provides the necessary functionality.
//
// TODO: securecookie redundantly applies two iterations of base64 encoding, increasing the output size (see https://github.com/gorilla/securecookie/issues/36).
// Since there's not that much code, we should consider modifying/implementing the encoding directly in this package.
type Codec struct {
	codecs []securecookie.Codec
}

func NewCodec(keyPairs [][]byte) *Codec {
	codecs := securecookie.CodecsFromPairs(keyPairs...)
	for _, c := range codecs {
		sc := c.(*securecookie.SecureCookie)
		sc.MaxLength(0)
		sc.SetSerializer(securecookie.GobEncoder{})
	}
	return &Codec{codecs: codecs}
}

func NewRandom() *Codec {
	keyPairs := [][]byte{securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32)}
	return NewCodec(keyPairs)
}

func (c *Codec) Encode(value any) (string, error) {
	return securecookie.EncodeMulti("", value, c.codecs...)
}

func (c *Codec) Decode(value string, dst any) error {
	return securecookie.DecodeMulti("", value, dst, c.codecs...)
}
