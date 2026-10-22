package kv

import (
	"crypto/sha256"
	"encoding/hex"
)

func (r *KV) hash(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	res := hasher.Sum(nil)

	return hex.EncodeToString(res)
}
