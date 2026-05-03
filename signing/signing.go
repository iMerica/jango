package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrBadSignature = errors.New("signing: signature does not match")
var ErrSignatureExpired = errors.New("signing: signature has expired")

type Signer struct {
	SecretKey string
	Sep       string
	Salt      string
}

func NewSigner(secretKey string, opts ...SignerOption) *Signer {
	s := &Signer{
		SecretKey: secretKey,
		Sep:       ":",
		Salt:      "",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type SignerOption func(*Signer)

func WithSep(sep string) SignerOption {
	return func(s *Signer) { s.Sep = sep }
}

func WithSalt(salt string) SignerOption {
	return func(s *Signer) { s.Salt = salt }
}

func (s *Signer) Signature(value string) string {
	key := s.deriveKey()
	h := hmac.New(sha256.New, key)
	h.Write([]byte(s.salted(value)))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Signer) Sign(value string) string {
	return value + s.Sep + s.Signature(value)
}

func (s *Signer) Unsign(signedValue string) (string, error) {
	idx := strings.LastIndex(signedValue, s.Sep)
	if idx < 0 {
		return "", ErrBadSignature
	}
	value := signedValue[:idx]
	sig := signedValue[idx+len(s.Sep):]
	expected := s.Signature(value)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", ErrBadSignature
	}
	return value, nil
}

func (s *Signer) deriveKey() []byte {
	h := sha256.New()
	h.Write([]byte(s.SecretKey))
	if s.Salt != "" {
		h.Write([]byte("signing.salt:" + s.Salt))
	}
	return h.Sum(nil)
}

func (s *Signer) salted(value string) string {
	if s.Salt != "" {
		return s.Salt + s.Sep + value
	}
	return value
}

type TimestampSigner struct {
	*Signer
}

func NewTimestampSigner(secretKey string, opts ...SignerOption) *TimestampSigner {
	return &TimestampSigner{
		Signer: NewSigner(secretKey, opts...),
	}
}

func (ts *TimestampSigner) Sign(value string) string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	return ts.Signer.Sign(value + ts.Sep + now)
}

func (ts *TimestampSigner) Unsign(value string, maxAge time.Duration) (string, error) {
	result, err := ts.Signer.Unsign(value)
	if err != nil {
		return "", err
	}
	idx := strings.LastIndex(result, ts.Sep)
	if idx < 0 {
		return "", ErrBadSignature
	}
	data := result[:idx]
	timestampStr := result[idx+len(ts.Sep):]
	var timestamp int64
	if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
		return "", ErrBadSignature
	}
	if maxAge > 0 {
		elapsed := time.Since(time.Unix(timestamp, 0))
		if elapsed > maxAge {
			return "", ErrSignatureExpired
		}
	}
	return data, nil
}

type TimestampSignerOption func(*TimestampSigner)

func b64Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func b64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func SignObject(secretKey string, obj interface{}, salt string) (string, error) {
	data, err := encodeObj(obj)
	if err != nil {
		return "", err
	}
	s := NewSigner(secretKey, WithSalt(salt))
	return s.Sign(b64Encode(data)), nil
}

func UnsignObject(secretKey string, signedValue string, salt string, dest interface{}) error {
	s := NewSigner(secretKey, WithSalt(salt))
	value, err := s.Unsign(signedValue)
	if err != nil {
		return err
	}
	data, err := b64Decode(value)
	if err != nil {
		return ErrBadSignature
	}
	return decodeObj(data, dest)
}

func SignObjectWithTimestamp(secretKey string, obj interface{}, salt string) (string, error) {
	data, err := encodeObj(obj)
	if err != nil {
		return "", err
	}
	ts := NewTimestampSigner(secretKey, WithSalt(salt))
	return ts.Sign(b64Encode(data)), nil
}

func UnsignObjectWithTimestamp(secretKey string, signedValue string, salt string, maxAge time.Duration, dest interface{}) error {
	ts := NewTimestampSigner(secretKey, WithSalt(salt))
	value, err := ts.Unsign(signedValue, maxAge)
	if err != nil {
		return err
	}
	data, err := b64Decode(value)
	if err != nil {
		return ErrBadSignature
	}
	return decodeObj(data, dest)
}

func encodeObj(obj interface{}) ([]byte, error) {
	switch v := obj.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return nil, fmt.Errorf("signing: unsupported object type %T", obj)
	}
}

func decodeObj(data []byte, dest interface{}) error {
	switch d := dest.(type) {
	case *string:
		*d = string(data)
		return nil
	case *[]byte:
		*d = data
		return nil
	default:
		return fmt.Errorf("signing: unsupported dest type %T", dest)
	}
}

func GetCookieSigner(secretKey string, salt string) *Signer {
	if salt == "" {
		salt = "jango.contrib.cookies"
	}
	return NewSigner(secretKey, WithSalt(salt))
}