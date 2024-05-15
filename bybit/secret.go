package bybit

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"maps"
	"time"
)

type SecretType string

const (
	SecretTypeRsa  = SecretType("RSA")
	SecretTypeHmac = SecretType("HMAC")
)

type SecretRecord struct {
	Type        SecretType `json:"type"` // "RSA" | "HMAC"
	ApiKey      string     `json:"apikey"`
	Secret      string     `json:"secret"`
	DateCreated time.Time  `json:"dateCreated"`
	DateExpired time.Time  `json:"dateExpired"`
}

func (r *SecretRecord) Expiry() time.Duration {
	return r.DateExpired.Sub(r.DateCreated)
}

func (r *SecretRecord) Hmac() ([]byte, error) {
	return []byte(r.Secret), nil
}

func (r *SecretRecord) Rsa() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(r.Secret))
	if block == nil {
		return nil, errors.New("PEM block not found")
	}

	some_key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	rsa_key, ok := some_key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected private key to be a RSA key")
	}

	return rsa_key, nil
}

func (r *SecretRecord) Sign(data []byte) (string, error) {
	switch r.Type {
	case "HMAC":
		return r.signByHmac(data)
	case "RSA":
		return r.signByRsa(data)
	default:
		return "", fmt.Errorf("unknown type of secret")
	}
}

func (r *SecretRecord) signByHmac(data []byte) (string, error) {
	key, err := r.Hmac()
	if err != nil {
		return "", err
	}

	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	signature := hex.EncodeToString(hash.Sum(nil))
	return signature, nil
}

func (r *SecretRecord) signByRsa(data []byte) (string, error) {
	key, err := r.Rsa()
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

type SecretStore map[uint64]SecretRecord

func (s SecretStore) Get(uid UserId) (SecretRecord, bool) {
	r, ok := s[uint64(uid)]
	return r, ok
}

func (s SecretStore) Set(uid UserId, record SecretRecord) {
	s[uint64(uid)] = record
}

func SaveSecrets(w io.Writer, records SecretStore) error {
	s := SecretStore{}
	for k, v := range records {
		v.Secret = base64.RawStdEncoding.EncodeToString([]byte(v.Secret))
		s[k] = v
	}

	data, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func LoadSecrets(r io.Reader, records SecretStore) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if len(data) == 0 {
		return nil
	}

	s := SecretStore{}
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for k, v := range s {
		r, err := base64.RawStdEncoding.DecodeString(v.Secret)
		if err != nil {
			return fmt.Errorf("%d has invalid secret: %w", k, err)
		}

		v.Secret = string(r)
		s[k] = v
	}

	maps.Copy(records, s)
	return nil
}
