package super

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	privKeyPath = ".keys/private.key"
	pubKeyPath  = ".keys/public.key"
)

func LoadOrCreateKeypair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}

		// Ensure directory exists
		if err := os.MkdirAll(".keys", 0700); err != nil {
			return nil, nil, fmt.Errorf("failed to create .keys directory: %w", err)
		}

		// Write private key
		if err := ioutil.WriteFile(privKeyPath, []byte(base64.StdEncoding.EncodeToString(priv)), 0600); err != nil {
			return nil, nil, fmt.Errorf("failed to write private key: %w", err)
		}

		// Write public key
		if err := ioutil.WriteFile(pubKeyPath, []byte(base64.StdEncoding.EncodeToString(pub)), 0644); err != nil {
			return nil, nil, fmt.Errorf("failed to write public key: %w", err)
		}

		fmt.Println("✅ Keys generated and saved to .keys/")
		return priv, pub, nil
	}

	// Read existing keys
	privEncoded, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	pubEncoded, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %w", err)
	}

	privDecoded, err := base64.StdEncoding.DecodeString(string(privEncoded))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	pubDecoded, err := base64.StdEncoding.DecodeString(string(pubEncoded))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	return ed25519.PrivateKey(privDecoded), ed25519.PublicKey(pubDecoded), nil
}

