package scenario

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug07_ClampCrossingCollapsesToMidpoint(t *testing.T) {
	got := Clamp(model.Range{Lo: -2, Hi: 2}, 3, 4); if got.Lo != 3 || got.Hi != 3 { t.Fatalf("%#v", got) }
}
