package file

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const utf8BOM = "\xEF\xBB\xBF"

type Frontmatter struct {
	Title          string
	Summary        string
	HasFrontmatter bool
}

type yamlFrontmatter struct {
	Title   string `yaml:"title"`
	Summary string `yaml:"summary"`
}

func ParseFrontmatter(content string) (Frontmatter, string) {
	fm := Frontmatter{}
	body := content

	trimmed := strings.TrimPrefix(content, utf8BOM)
	if !strings.HasPrefix(trimmed, "---") {
		return fm, body
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return fm, body
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return fm, body
	}

	yamlText := strings.Join(lines[1:endIdx], "\n")
	var yf yamlFrontmatter
	if err := yaml.Unmarshal([]byte(yamlText), &yf); err != nil {
		return fm, body
	}

	fm.Title = yf.Title
	fm.Summary = yf.Summary
	fm.HasFrontmatter = true
	body = strings.Join(lines[endIdx+1:], "\n")
	body = strings.TrimPrefix(body, "\n")
	return fm, body
}
