package validation

import (
	"testing"

	"task177-isomix/internal/model"
)

// SameDimensions must compare exact isotope names, not only cardinality.
func TestSameDimensionsExactNames(t *testing.T) {
	a := map[string]model.Range{"d18O": {}, "d2H": {}}
	// 同名同基数 -> 相等。
	if !SameDimensions(a, map[string]model.Range{"d18O": {}, "d2H": {}}) {
		t.Fatalf("SameDimensions should return true for identical name sets")
	}
	// 同基数但名字不同 -> 不应相等。
	if SameDimensions(a, map[string]model.Range{"d13C": {}, "d34S": {}}) {
		t.Fatalf("SameDimensions should return false for same cardinality but different names")
	}
	// 基数不同 -> 不应相等。
	if SameDimensions(a, map[string]model.Range{"d18O": {}}) {
		t.Fatalf("SameDimensions should return false for different cardinality")
	}
	// 空映射与空映射相等。
	if !SameDimensions(map[string]model.Range{}, map[string]model.Range{}) {
		t.Fatalf("SameDimensions should return true for two empty maps")
	}
}
