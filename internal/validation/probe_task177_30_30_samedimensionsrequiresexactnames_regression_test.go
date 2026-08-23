package validation

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug30_SameDimensionsRequiresExactNames(t *testing.T) {
	a := map[string]model.Range{"d18O": {}}; b := map[string]model.Range{"d2H": {}}; if SameDimensions(a, b) { t.Fatalf("different dimensions accepted") }
}
