package commons

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"puma-gallery/db"
	"strings"

	"context"
	"database/sql"
	_ "embed"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CompareHashToPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(tokenLength int) (string, error) {
	bytes := make([]byte, tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	} else {
		return base64.URLEncoding.EncodeToString(bytes), nil
	}
}

//go:embed schema.sql
var ddl string

func CreateNewDb() (*sql.DB, error) {
	ctx := context.Background()

	db, err := sql.Open("sqlite", "users.db")
	if err != nil {
		return nil, err
	}

	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return nil, err
	}

	return db, nil
}

var ErrUnauthorized = errors.New("unauthorized")

func AuthorizePost(c *fiber.Ctx) error {
	sqlDb, err := CreateNewDb()
	if err != nil {
		return ErrUnauthorized
	}
	st := c.Cookies("session_token", "")
	if st == "" {
		return ErrUnauthorized
	}
	queries := db.New(sqlDb)
	ctx := context.Background()
	user, err := queries.GetUserBySessionToken(ctx, sql.NullString{String: st, Valid: true})
	if err != nil {
		return ErrUnauthorized
	}
	csrf := c.Cookies("csrf_token", "")
	if csrf == "" {
		return ErrUnauthorized
	}
	if csrf != user.CsrfToken.String {
		return ErrUnauthorized
	}
	return nil
}

func AuthorizeGet(c *fiber.Ctx) error {
	sqlDb, err := CreateNewDb()
	if err != nil {
		return ErrUnauthorized
	}
	st := c.Cookies("session_token", "")
	if st == "" {
		return ErrUnauthorized
	}
	queries := db.New(sqlDb)
	ctx := context.Background()
	_, err = queries.GetUserBySessionToken(ctx, sql.NullString{String: st, Valid: true})
	if err != nil {
		return ErrUnauthorized
	}
	return nil
}

type ImageData struct {
	Caption   string  `json:"caption"`
	CreatedAt *string `json:"created_at"`
	FilePath  string  `json:"filePath"`
	Id        int     `json:"id"`
	Url       string  `json:"url"`
}

type ImageToUpload struct {
	Caption  string `json:"caption"`
	FilePath string `json:"filePath"`
	Url      string `json:"url"`
}

type KafkaImage struct {
	Url    string `json:"image_url"`
	ApiKey string `json:"api_key"`
}

type KafkaData struct {
	JsonPayload []byte `json:"json_payload"`
	AesKey      []byte `json:"aes_key"`
}

type KafkaSend struct {
	Value string `json:"value"`
}

func LoadEncryptionKeyFromEnv() string {
	key := os.Getenv("E2E_PUBLIC_KEY")
	return strings.ReplaceAll(key, "\\n", "\n")
}

func StringKeyToRsaKey(key string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func EncryptDataRsa(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		publicKey,
		data,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func EncryptDataAes(data []byte) ([]byte, []byte, error) {
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nil, nonce, data, nil)
	encryptedDataWithNonce := append(nonce, encryptedData...)
	rsaKey, err := StringKeyToRsaKey(LoadEncryptionKeyFromEnv())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get RSA key: %w", err)
	}
	encryptedKey, err := EncryptDataRsa(rsaKey, aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}
	return encryptedDataWithNonce, encryptedKey, nil
}
