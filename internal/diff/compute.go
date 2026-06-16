package diff

import (
	"github.com/yiwocapital/yiwo-draft-viewer/internal/model"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func Compute(a, b string) []model.DiffSegment {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, true)
	diffs = dmp.DiffCleanupSemantic(diffs)

	out := make([]model.DiffSegment, 0, len(diffs))
	for _, d := range diffs {
		op := model.DiffEqual
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			op = model.DiffInsert
		case diffmatchpatch.DiffDelete:
			op = model.DiffDelete
		}
		out = append(out, model.DiffSegment{Op: op, Text: d.Text})
	}
	return out
}
