package scenario

import (
	"testing"
	"task177-isomix/internal/model"
)

func TestBug06_TransformScalesWidthAroundCenter(t *testing.T) {
	got := Transform(model.Range{Lo: 2, Hi: 6}, .5, 0); if got.Lo != 3 || got.Hi != 5 { t.Fatalf("%#v", got) }
}
