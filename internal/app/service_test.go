package app

import (
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