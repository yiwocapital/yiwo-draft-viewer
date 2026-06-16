package file

import (
	"strings"
	"testing"
)

func TestParseFrontmatter_Full(t *testing.T) {
	content := `---
title: 测试标题
summary: 这里是摘要
---
# 正文开始
这是正文。`
	meta, body := ParseFrontmatter(content)
	if meta.Title != "测试标题" {
		t.Errorf("expected title '测试标题', got %q", meta.Title)
	}
	if meta.Summary != "这里是摘要" {
		t.Errorf("expected summary, got %q", meta.Summary)
	}
	if !meta.HasFrontmatter {
		t.Error("expected HasFrontmatter=true")
	}
	if !strings.Contains(body, "# 正文开始") {
		t.Error("body should contain rest of content")
	}
}

func TestParseFrontmatter_TitleOnly(t *testing.T) {
	content := `---
title: 只有标题
---
正文`
	meta, _ := ParseFrontmatter(content)
	if meta.Title != "只有标题" {
		t.Errorf("expected title, got %q", meta.Title)
	}
	if meta.Summary != "" {
		t.Errorf("expected empty summary, got %q", meta.Summary)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# 标题\n正文"
	meta, body := ParseFrontmatter(content)
	if meta.HasFrontmatter {
		t.Error("expected HasFrontmatter=false")
	}
	if body != content {
		t.Errorf("body should equal full content when no frontmatter")
	}
}

func TestParseFrontmatter_BrokenYAML(t *testing.T) {
	content := `---
title: [invalid
: not yaml
---
正文`
	meta, _ := ParseFrontmatter(content)
	if meta.HasFrontmatter {
		t.Error("expected HasFrontmatter=false on broken YAML")
	}
}

func TestParseFrontmatter_BOM(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	content := bom + "---\ntitle: 有BOM\n---\n正文"
	meta, _ := ParseFrontmatter(content)
	if meta.Title != "有BOM" {
		t.Errorf("expected title '有BOM', got %q", meta.Title)
	}
}
