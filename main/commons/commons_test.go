package commons

import (
	"os"
	"slices"
	"testing"
)

func TestHashingUtils(t *testing.T) {
	testCases := []struct {
		passwordOne    string
		passwordTwo    string
		passwordsMatch bool
	}{
		{"hello", "hello", true},
		{"hello", "bye", false},
	}
	for _, tc := range testCases {
		hashedOne, err := HashPassword(tc.passwordOne)
		if err != nil {
			t.Errorf("Not expecting any error while hashing, got %s", err.Error())
		} else {
			comparison := CompareHashToPassword(tc.passwordTwo, hashedOne)
			if tc.passwordsMatch != comparison {
				t.Errorf("Expecting password-to-hash comparison to yield %v, got %v", tc.passwordsMatch, comparison)
			}
		}
	}
}

func TestCreateDb(t *testing.T) {
	_, err := CreateNewDb()
	if err != nil {
		t.Errorf("Not expecting an error when creating a new database instance, got %s", err.Error())
	}
}

func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		tokenLength          int
		expectedStringLength int
	}{
		{32, 44},
		{16, 24},
		{48, 64},
	}
	for _, tc := range testCases {
		token, err := GenerateToken(tc.tokenLength)
		if err != nil {
			t.Errorf("Not expecting an error when generating a new token, got %s", err.Error())
		}
		if len(token) != tc.expectedStringLength {
			t.Errorf("Expecting base64-encoded token to be of length %d, got %d", tc.expectedStringLength, len(token))
		}
	}
}

func TestEncryptionUtils(t *testing.T) {
	if _, ok := os.LookupEnv("E2E_PUBLIC_KEY"); !ok {
		t.Skip("`E2E_PUBLIC_KEY` not set as an environment variable")
	}
	key, err := LoadEncryptionKeyFromEnv()
	if err != nil {
		t.Errorf("Expecting no error while loading the encryption key from environment, got %s", err.Error())
	} else {
		pk, err := StringKeyToRsaKey(key)
		if err != nil {
			t.Errorf("Expecting no error while converted the loaded key to an RSA key, got %s", err.Error())
		} else {
			dt, err := EncryptDataRsa(pk, []byte("hello world"))
			if err != nil {
				t.Errorf("Expecting no error while encrypting data with RSA, got %s", err.Error())
			} else {
				enc, _, err := EncryptDataAes([]byte("hello world"))
				if err != nil {
					t.Errorf("Expecting no error while encrypting data with AES, got %s", err.Error())
				} else {
					if slices.Equal(dt, enc) {
						t.Error("Expecting different encryption algorithms to produce different results, but they are the same.")
					}
				}
			}
		}
	}
}
