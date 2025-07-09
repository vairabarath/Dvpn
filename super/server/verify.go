package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
)

func verifyClientPeer(
	peerID string,
	region string,
	os string,
	natType string,
	nonce string,
	pubKeyBase64 string,
	signatureBase64 string,
) bool {
	// 1. Rebuild the signed message
	msg := fmt.Sprintf("%s|%s|%s|%s|%s", peerID, region, os, natType, nonce)

	// 2. Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		log.Printf("❌ Failed to decode public key: %v", err)
		return false
	}

	// 3. Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		log.Printf("❌ Failed to decode signature: %v", err)
		return false
	}

	// 4. Verify signature
	valid := ed25519.Verify(ed25519.PublicKey(pubKeyBytes), []byte(msg), sigBytes)
	if !valid {
		log.Printf("❌ Invalid signature for peer %s", peerID)
	} else {
		log.Printf("✅ Signature verified for peer %s", peerID)
	}
	return valid
}
