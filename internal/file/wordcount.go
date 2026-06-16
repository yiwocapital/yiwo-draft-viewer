package file

import (
	"unicode"
	"unicode/utf8"
)

func CountChars(text string) int {
	if text == "" {
		return 0
	}
	inSpace := false
	count := 0
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size
		if unicode.IsSpace(r) {
			if !inSpace {
				count++
				inSpace = true
			}
		} else {
			count++
			inSpace = false
		}
	}
	return count
}
