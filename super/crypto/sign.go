package super

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func SignPayload(priv ed25519.PrivateKey, id, region, ip, nonce string) (string) {
	msg := fmt.Sprintf("%s|%s|%s|%s", id, region, ip, nonce)
	sign := ed25519.Sign(priv, []byte(msg))
	return base64.StdEncoding.EncodeToString(sign)
}

func GenerateNonce() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic("Failed to generate nonce: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(b)
}