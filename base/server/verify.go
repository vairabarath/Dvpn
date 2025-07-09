package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
)


func VerifySuperNodeSignature(nodeID, region, ip, nonce, pubKeyBase64, signatureBase64 string) bool {
	msg := fmt.Sprintf("%s|%s|%s|%s", nodeID, region, ip, nonce)
	
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return false
	}

	signBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	return ed25519.Verify(pubKeyBytes, []byte(msg), signBytes)
}