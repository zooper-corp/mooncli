package tools

import (
	"strings"
	"unicode"
)

func ToAscii(s string) string {
	t := strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return -1
		}
		return r
	}, s)
	return t
}
