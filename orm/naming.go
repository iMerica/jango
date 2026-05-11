package orm

import "strings"

// GoFieldToDBColumn converts exported Go field names to Django-style
// PostgreSQL column names.
func GoFieldToDBColumn(name string) string {
	if name == "" {
		return ""
	}

	var b strings.Builder
	runes := []rune(name)
	for i, r := range runes {
		if i > 0 && shouldInsertUnderscore(runes, i) {
			b.WriteRune('_')
		}
		b.WriteRune(toLowerASCII(r))
	}
	return b.String()
}

func shouldInsertUnderscore(runes []rune, i int) bool {
	curr := runes[i]
	prev := runes[i-1]
	if !isUpperASCII(curr) {
		return false
	}
	if isLowerASCII(prev) || isDigitASCII(prev) {
		return true
	}
	return isUpperASCII(prev) && i+1 < len(runes) && isLowerASCII(runes[i+1])
}

func toLowerASCII(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

func isUpperASCII(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isLowerASCII(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isDigitASCII(r rune) bool {
	return r >= '0' && r <= '9'
}
