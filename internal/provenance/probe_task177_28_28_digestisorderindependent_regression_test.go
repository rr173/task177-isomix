package provenance

import (
	"testing"
)

func TestBug28_DigestIsOrderIndependent(t *testing.T) {
	events := []Event{{Kind: "endmember", ID: "b", Version: 1}, {Kind: "endmember", ID: "a", Version: 1}}; if Digest(events) != Digest([]Event{events[1], events[0]}) { t.Fatalf("digest changed") }
}
