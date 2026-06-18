package app

import (
	"testing"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/model"
)

func TestSplitOutComments_ModifiedComment(t *testing.T) {
	// Simulates the diff-match-patch output when a comment is modified:
	// <!-- comment v1 --> → <!-- comment v2 -->
	// Each fragment of the comment sits in a different segment. The
	// previous regex-per-segment approach failed to detect the comment
	// because no single segment contains both <!-- and -->. The new
	// character-stream approach matches by reconstructing the full text.
	segs := []model.DiffSegment{
		{Op: model.DiffEqual, Text: "上文\n<!-- comment v"},
		{Op: model.DiffDelete, Text: "1"},
		{Op: model.DiffInsert, Text: "2"},
		{Op: model.DiffEqual, Text: " -->\nmore text"},
	}
	out := splitOutComments(segs)

	// Collect comment segments and their concatenated text. On the right
	// side, the comment reads `<!-- comment v2 -->`; on the left side it
	// reads `<!-- comment v1 -->`. Across the stream the IsComment-flagged
	// chars join to `<!-- comment v12 -->` (the inserted "2" sits between
	// the equal `<!-- comment v` and ` -->`, and the deleted "1" does too
	// in the left-side match). The point is: every character that lives
	// inside *some* version of the comment is marked IsComment, even
	// though no single segment contains the full regex match.
	var commentSegs []model.DiffSegment
	for _, s := range out {
		if s.IsComment {
			commentSegs = append(commentSegs, s)
		}
	}
	if len(commentSegs) == 0 {
		t.Fatalf("expected at least one IsComment segment; got: %+v", out)
	}
	var joined string
	for _, s := range commentSegs {
		joined += s.Text
	}
	// All four cross-segment fragments must be marked as comment.
	if joined != "<!-- comment v12 -->" {
		t.Errorf("joined comment text = %q, want %q", joined, "<!-- comment v12 -->")
	}
	// The comment segments should keep the op of their source character
	// (Delete for "1", Insert for "2", Equal for the rest).
	hasDelete := false
	hasInsert := false
	hasEqual := false
	for _, s := range commentSegs {
		switch s.Op {
		case model.DiffDelete:
			hasDelete = true
		case model.DiffInsert:
			hasInsert = true
		case model.DiffEqual:
			hasEqual = true
		}
	}
	if !hasDelete || !hasInsert || !hasEqual {
		t.Errorf("expected all three ops in comment segments; got delete=%v insert=%v equal=%v", hasDelete, hasInsert, hasEqual)
	}
}

func TestSplitOutComments_UnchangedComment(t *testing.T) {
	segs := []model.DiffSegment{
		{Op: model.DiffEqual, Text: "上文\n<!-- unchanged -->\nmore text"},
	}
	out := splitOutComments(segs)
	var commentSegs []model.DiffSegment
	for _, s := range out {
		if s.IsComment {
			commentSegs = append(commentSegs, s)
		}
	}
	if len(commentSegs) != 1 {
		t.Fatalf("expected 1 comment segment, got %d", len(commentSegs))
	}
	if commentSegs[0].Text != "<!-- unchanged -->" {
		t.Errorf("comment text = %q, want %q", commentSegs[0].Text, "<!-- unchanged -->")
	}
	if commentSegs[0].Op != model.DiffEqual {
		t.Errorf("op = %d, want DiffEqual", commentSegs[0].Op)
	}
}
