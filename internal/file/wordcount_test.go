package file

import "testing"

func TestCountChars_Chinese(t *testing.T) {
	n := CountChars("你好世界")
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
}

func TestCountChars_CollapsesWhitespace(t *testing.T) {
	n := CountChars("你好\n\n\n   世界")
	if n != 5 { // "你好" + 1 collapsed space + "世界" = 5
		t.Errorf("expected 5, got %d", n)
	}
}

func TestCountChars_Punctuation(t *testing.T) {
	n := CountChars("你好，世界！")
	if n != 6 {
		t.Errorf("expected 6, got %d", n)
	}
}

func TestCountChars_Empty(t *testing.T) {
	n := CountChars("")
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestCountChars_Mixed(t *testing.T) {
	n := CountChars("Hello 世界!")
	// "Hello"(5) + 1 space + "世界"(2) + "!"(1) = 9
	if n != 9 {
		t.Errorf("expected 9, got %d", n)
	}
}

func TestCountChars_StripsComments(t *testing.T) {
	text := "正文文字<!-- 这是编辑注释，不应计入字数 -->更多正文"
	n := CountChars(text)
	// After stripping the comment block, the result is "正文文字更多正文"
	// (the spaces flanking the comment are also removed as part of the strip).
	// 8 non-space chars, 0 spaces = 8.
	if n != 8 {
		t.Errorf("expected 8, got %d", n)
	}
}

func TestCountChars_StripsMultilineComments(t *testing.T) {
	text := "上文\n<!--\n多行注释\n第二行\n-->\n下文"
	n := CountChars(text)
	// After stripping: "上文\n下文" — 4 chars + 1 newline (collapsed to space) = 5.
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
}
