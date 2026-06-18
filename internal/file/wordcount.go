package file

import (
	"regexp"
	"unicode"
	"unicode/utf8"
)

// htmlCommentRegex matches HTML comments: <!-- ... -->
// (?s) makes . match newlines so multi-line comments are removed as one block.
var htmlCommentRegex = regexp.MustCompile(`(?s)<!--.*?-->`)

// StripComments removes all HTML comment blocks from the text.
func StripComments(text string) string {
	return htmlCommentRegex.ReplaceAllString(text, "")
}

func CountChars(text string) int {
	if text == "" {
		return 0
	}
	text = StripComments(text)
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
