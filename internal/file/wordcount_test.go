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
