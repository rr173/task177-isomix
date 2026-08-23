package scenario

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug08_CompareIncludesMissingSourceAsZeroRange(t *testing.T) {
	rows := Compare(map[string]model.Range{"a": {Lo: 0, Hi: 1}}, map[string]model.Range{"b": {Lo: 0, Hi: 2}}); if len(rows) != 2 { t.Fatalf("rows=%d", len(rows)) }
}
