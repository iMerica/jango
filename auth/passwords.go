package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/bcrypt"
)

const (
	PBKDF2Algorithm  = "pbkdf2_sha256"
	PBKDF2Iterations = 390000
	PBKDF2SaltLength = 12
	PBKDF2HashLength = 32

	BcryptAlgorithm = "bcrypt"
	BcryptCost      = 12

	SHA1Algorithm = "sha1"
)

type PasswordHasher interface {
	Algorithm() string
	Hash(password string) (string, error)
	Verify(password, encoded string) bool
	MustUpdate(encoded string) bool
}

var hashers []PasswordHasher

func init() {
	hashers = []PasswordHasher{
		&PBKDF2PasswordHasher{Iterations: PBKDF2Iterations},
		&BcryptPasswordHasher{Cost: BcryptCost},
	}
}

func RegisterHasher(h PasswordHasher) {
	hashers = append(hashers, h)
}

func SetHashers(list []PasswordHasher) {
	hashers = list
}

func GetHashers() []PasswordHasher {
	return hashers
}

func MakePassword(password string) (string, error) {
	if len(hashers) == 0 {
		return "", fmt.Errorf("auth: no password hashers registered")
	}
	return hashers[0].Hash(password)
}

func CheckPassword(password, encoded string) bool {
	if encoded == "" {
		return false
	}
	parts := strings.SplitN(encoded, "$", 2)
	if len(parts) != 2 {
		return false
	}
	algo := parts[0]

	for _, h := range hashers {
		if h.Algorithm() == algo {
			return h.Verify(password, encoded)
		}
	}

	if algo == "pbkdf2_sha256" {
		h := &PBKDF2PasswordHasher{Iterations: PBKDF2Iterations}
		return h.Verify(password, encoded)
	}

	return false
}

func MustUpdatePassword(encoded string) bool {
	parts := strings.SplitN(encoded, "$", 2)
	if len(parts) != 2 {
		return false
	}
	algo := parts[0]
	for _, h := range hashers {
		if h.Algorithm() == algo {
			return h.MustUpdate(encoded)
		}
	}
	return true
}

func GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(salt)[:length], nil
}

type PBKDF2PasswordHasher struct {
	Iterations int
}

func (h *PBKDF2PasswordHasher) Algorithm() string { return PBKDF2Algorithm }

func (h *PBKDF2PasswordHasher) Hash(password string) (string, error) {
	salt, err := GenerateSalt(PBKDF2SaltLength)
	if err != nil {
		return "", err
	}
	dk := pbkdf2.Key([]byte(password), []byte(salt), h.Iterations, PBKDF2HashLength, sha256.New)
	encoded := base64.RawStdEncoding.EncodeToString(dk)
	return fmt.Sprintf("%s$%d$%s$%s", PBKDF2Algorithm, h.Iterations, salt, encoded), nil
}

func (h *PBKDF2PasswordHasher) Verify(password, encoded string) bool {
	parts := strings.SplitN(encoded, "$", 4)
	if len(parts) != 4 || parts[0] != PBKDF2Algorithm {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	salt := parts[2]
	storedHash := parts[3]

	dk := pbkdf2.Key([]byte(password), []byte(salt), iterations, PBKDF2HashLength, sha256.New)
	candidate := base64.RawStdEncoding.EncodeToString(dk)

	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(candidate)) == 1
}

func (h *PBKDF2PasswordHasher) MustUpdate(encoded string) bool {
	parts := strings.SplitN(encoded, "$", 4)
	if len(parts) < 2 {
		return true
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}
	return iterations != h.Iterations
}

type BcryptPasswordHasher struct {
	Cost int
}

func (h *BcryptPasswordHasher) Algorithm() string { return BcryptAlgorithm }

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = BcryptCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return BcryptAlgorithm + "$" + string(hash), nil
}

func (h *BcryptPasswordHasher) Verify(password, encoded string) bool {
	parts := strings.SplitN(encoded, "$", 2)
	if len(parts) != 2 || parts[0] != BcryptAlgorithm {
		return false
	}
	hash := parts[1]
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (h *BcryptPasswordHasher) MustUpdate(encoded string) bool {
	parts := strings.SplitN(encoded, "$", 2)
	if len(parts) != 2 {
		return true
	}
	hash := parts[1]
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true
	}
	targetCost := h.Cost
	if targetCost == 0 {
		targetCost = BcryptCost
	}
	return cost != targetCost
}

func checkPasswordSHA1(password, encoded string) bool {
	dk := sha256.Sum256([]byte(password))
	computed := fmt.Sprintf("%x", dk)
	parts := strings.SplitN(encoded, "$", 2)
	if len(parts) != 2 || parts[0] != SHA1Algorithm {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(parts[1]), []byte(computed)) == 1
}

func MakeRandomPassword(length int) (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b), nil
}