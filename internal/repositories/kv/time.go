package kv

import "time"

func (k *KV) now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
