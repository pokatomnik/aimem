package qdrant

import "time"

// now renders a creation timestamp the same way kv does, so facts stored by
// either backend read back identically.
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
