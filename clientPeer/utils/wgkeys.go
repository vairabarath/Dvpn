// utils/wgkeys.go
package utils

import (
	"encoding/base64"
	"os"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	wgPrivPath = ".keys/wg_private.key"
	wgPubPath  = ".keys/wg_public.key"
)

func LoadOrCreateWGKeypair() (wgtypes.Key, wgtypes.Key, error) {
	if _, err := os.Stat(wgPrivPath); os.IsNotExist(err) {
		priv, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return wgtypes.Key{}, wgtypes.Key{}, err
		}
		pub := priv.PublicKey()

		_ = os.MkdirAll(".keys", 0700)
		_ = os.WriteFile(wgPrivPath, []byte(priv.String()), 0600)
		_ = os.WriteFile(wgPubPath, []byte(pub.String()), 0644)

		return priv, pub, nil
	}

	privBytes, _ := os.ReadFile(wgPrivPath)
	pubBytes, _ := os.ReadFile(wgPubPath)

	priv, _ := wgtypes.ParseKey(string(privBytes))
	pub, _ := wgtypes.ParseKey(string(pubBytes))

	return priv, pub, nil
}

func PublicKeyBase64(pub wgtypes.Key) string {
	return base64.StdEncoding.EncodeToString([]byte(pub.String()))
}
