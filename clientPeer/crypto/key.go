package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	privKeyPath = ".keys/client_private.key"
	pubKeyPath  = ".keys/client_public.key"
)

func LoadOrCreateKeypair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}

		_ = os.MkdirAll(".keys", 0700)
		_ = ioutil.WriteFile(privKeyPath, []byte(base64.StdEncoding.EncodeToString(priv)), 0600)
		_ = ioutil.WriteFile(pubKeyPath, []byte(base64.StdEncoding.EncodeToString(pub)), 0644)

		fmt.Println("✅ Keys generated and saved to .keys/")
		return priv, pub, nil
	}

	privEncoded, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	pubEncoded, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read public key: %w", err)
	}

	privDecoded, _ := base64.StdEncoding.DecodeString(string(privEncoded))
	pubDecoded, _ := base64.StdEncoding.DecodeString(string(pubEncoded))

	return ed25519.PrivateKey(privDecoded), ed25519.PublicKey(pubDecoded), nil
}

func GenerateNonce() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic("Failed to generate nonce: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(b)
}
func SignPeerPayload(priv ed25519.PrivateKey, id, region, os, nat, nonce string) string {
	msg := fmt.Sprintf("%s|%s|%s|%s|%s", id, region, os, nat, nonce)
	sign := ed25519.Sign(priv, []byte(msg))
	return base64.StdEncoding.EncodeToString(sign)
}
