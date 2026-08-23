package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Digest hashes the ordered event keys, independent of map iteration order.
func Digest(events []Event) string {
	keys := make([]string, 0, len(events))
	for _, event := range events {
		keys = append(keys, eventKey(event))
	}
	sort.Strings(keys)
	sum := sha256.Sum256([]byte(strings.Join(keys, "\n")))
	return hex.EncodeToString(sum[:])
}
