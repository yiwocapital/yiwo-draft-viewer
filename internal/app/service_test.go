package app

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenFile_ReturnsLoadResult(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `---
title: 测试
summary: 摘要
---
# 正文`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)

	res := s.OpenFile(path)
	if !res.Ok {
		t.Fatalf("expected ok, got error: %s", res.Error)
	}
	data, ok := res.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["title"] != "测试" {
		t.Errorf("expected title '测试', got %v", data["title"])
	}
	if data["summary"] != "摘要" {
		t.Errorf("expected summary '摘要', got %v", data["summary"])
	}
}

func TestOpenFile_RejectsOversize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "big.md")
	big := strings.Repeat("a", 6*1024*1024)
	if err := os.WriteFile(path, []byte(big), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)

	res := s.OpenFile(path)
	if res.Ok {
		t.Fatal("expected failure for oversize file")
	}
	if !strings.Contains(res.Error, "too large") {
		t.Errorf("expected 'too large' error, got %q", res.Error)
	}
}

func TestCopySection_All(t *testing.T) {
	s := NewService()
	s.content = "---\ntitle: T\n---\nbody"
	s.title = "T"
	s.summary = ""
	s.hasFm = true
	s.body = "body"

	res := s.CopySection("all")
	if !res.Ok {
		t.Fatalf("expected ok, got %s", res.Error)
	}
	if data := res.Data.(map[string]interface{}); data["text"] != "---\ntitle: T\n---\nbody" {
		t.Errorf("unexpected text: %v", data["text"])
	}
}

func TestCopySection_Title(t *testing.T) {
	s := NewService()
	s.content = "---\ntitle: 我的标题\n---\nbody"
	s.title = "我的标题"
	s.hasFm = true
	s.body = "body"

	res := s.CopySection("title")
	if !res.Ok {
		t.Fatalf("expected ok, got %s", res.Error)
	}
	if data := res.Data.(map[string]interface{}); data["text"] != "我的标题" {
		t.Errorf("expected '我的标题', got %v", data["text"])
	}
}

func TestCopySection_NoFrontmatter(t *testing.T) {
	s := NewService()
	s.hasFm = false

	res := s.CopySection("title")
	if res.Ok {
		t.Fatal("expected failure when no frontmatter")
	}
}

func TestCleanForCopy(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no comments or extra blanks",
			in:   "para1\npara2",
			want: "para1\npara2",
		},
		{
			name: "strips multi-line comment",
			in:   "before\n<!--\n  long editor note\n  spans lines\n-->\nafter",
			want: "before\n\nafter",
		},
		{
			name: "strips inline comment",
			in:   "text <!-- inline --> more",
			want: "text  more",
		},
		{
			name: "collapses 3+ newlines to 2",
			in:   "a\n\n\n\nb",
			want: "a\n\nb",
		},
		{
			name: "preserves single blank line",
			in:   "a\n\nb",
			want: "a\n\nb",
		},
		{
			name: "removes leading/trailing blank lines",
			in:   "\n\n\na\nb\n\n\n",
			want: "a\nb",
		},
		{
			name: "combines comment + extra blanks",
			in:   "a\n\n\n\n<!-- note -->\nb",
			want: "a\n\nb",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cleanForCopy(tc.in)
			if got != tc.want {
				t.Errorf("cleanForCopy(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)
	s.OpenFile(path)
	s.editStartHash = sha256Hex("original")

	res := s.Save("modified content")
	if !res.Ok {
		t.Fatalf("expected ok, got error: %s", res.Error)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "modified content" {
		t.Errorf("file content: got %q, want %q", got, "modified content")
	}

	// tmp file must be cleaned up
	if _, err := os.Stat(path + ".yiwo-tmp"); !os.IsNotExist(err) {
		t.Errorf("tmp file should not exist after Save")
	}
}

func TestSave_ExternalModified(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)
	s.OpenFile(path)
	s.editStartHash = sha256Hex("original")

	// Simulate external modification
	if err := os.WriteFile(path, []byte("changed externally"), 0644); err != nil {
		t.Fatal(err)
	}

	res := s.Save("my edit")
	if res.Ok {
		t.Fatal("expected failure due to external modification")
	}
	if res.Code != "EXTERNAL_MODIFIED" {
		t.Errorf("expected code EXTERNAL_MODIFIED, got %q", res.Code)
	}

	// File should not have been overwritten
	got, _ := os.ReadFile(path)
	if string(got) != "changed externally" {
		t.Errorf("file should be unchanged after conflict, got %q", got)
	}
}

func TestSave_DirtyCleared(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)
	s.OpenFile(path)
	s.editStartHash = sha256Hex("original")
	s.dirty = true

	s.Save("modified")

	if s.dirty {
		t.Error("expected dirty=false after successful Save")
	}
}

func TestSave_OverwritesExternalWhenForced(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	s := NewService()
	s.SetConfigDir(dir)
	s.OpenFile(path)
	s.editStartHash = sha256Hex("original")

	// Simulate external modification
	os.WriteFile(path, []byte("changed"), 0644)

	res := s.SaveOverwrite("forced write")
	if !res.Ok {
		t.Fatalf("expected ok, got error: %s", res.Error)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "forced write" {
		t.Errorf("file should be overwritten, got %q", got)
	}
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}