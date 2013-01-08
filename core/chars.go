package din

func isLineEnding(r rune) bool {
	switch r {
	case '\n', '\r':
		return true
	}
	return false
}

func isWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

func isAlpha(r rune) bool {
	return isLowerAlpha(r) || isUpperAlpha(r)
}

func isLowerAlpha(r rune) bool {
	i := int(r)
	return 97 <= i && i <= 122
}

func isUpperAlpha(r rune) bool {
	i := int(r)
	return 65 <= i && i <= 90
}

func isNum(r rune) bool {
	i := int(r)
	return 48 <= i && i <= 57
}

func isAlphaNum(r rune) bool {
	return isNum(r) || isAlpha(r)
}
