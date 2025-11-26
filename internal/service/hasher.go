package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignData(value []byte, key string) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write(value)

	return hex.EncodeToString(hasher.Sum(nil))
}
