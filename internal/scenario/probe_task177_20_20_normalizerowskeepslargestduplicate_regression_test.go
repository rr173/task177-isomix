package scenario

import (
	"testing"
)

func TestBug20_NormalizeRowsKeepsLargestDuplicate(t *testing.T) {
	rows := NormalizeRows([]Projection{{ID: "x", DeltaWidth: .1}, {ID: "x", DeltaWidth: -.4}}); if len(rows) != 1 || rows[0].DeltaWidth != -.4 { t.Fatalf("%#v", rows) }
}
