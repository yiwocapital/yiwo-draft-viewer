package diff

import (
	"testing"

	"github.com/yiwocapital/yiwo-draft-viewer/internal/model"
)

func TestCompute_InsertsAtEnd(t *testing.T) {
	segs := Compute("hello", "hello world")
	if segs[len(segs)-1].Op != model.DiffInsert || segs[len(segs)-1].Text != " world" {
		t.Errorf("expected trailing insert of ' world', got %+v", segs)
	}
}

func TestCompute_DeletesInMiddle(t *testing.T) {
	segs := Compute("hello cruel world", "hello world")
	found := false
	for _, s := range segs {
		if s.Op == model.DiffDelete && (s.Text == " cruel" || s.Text == "cruel " || s.Text == " cruel ") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected delete of ' cruel' (with adjacent space), got %+v", segs)
	}
}

func TestCompute_EmptyLeft(t *testing.T) {
	segs := Compute("", "hello")
	all := ""
	for _, s := range segs {
		if s.Op == model.DiffInsert {
			all += s.Text
		}
	}
	if all != "hello" {
		t.Errorf("expected all-insert 'hello', got %q", all)
	}
}

func TestCompute_Equal(t *testing.T) {
	segs := Compute("same", "same")
	if len(segs) != 1 || segs[0].Op != model.DiffEqual {
		t.Errorf("expected single equal segment, got %+v", segs)
	}
}

func TestCompute_Chinese(t *testing.T) {
	segs := Compute("我喜欢苹果", "我喜欢苹果和梨")
	hasInsert := false
	for _, s := range segs {
		if s.Op == model.DiffInsert && s.Text == "和梨" {
			hasInsert = true
		}
	}
	if !hasInsert {
		t.Errorf("expected Chinese insert '和梨', got %+v", segs)
	}
}
