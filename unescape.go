package iislogs

func unescape(s string) (ret string) {
	inEscape := 0
	escapedChar := byte(0)
	for i := 0; i < len(s); i++ {
		next := byte(0)
		switch {
		case inEscape == 0 && s[i] != '%':
			next = s[i]
		case inEscape > 0 && s[i] == '%':
			escapedChar = 0
			inEscape = 0
			next = s[i]
		case inEscape == 0 && s[i] == '%':
			inEscape = 1
			continue
		case inEscape > 0:
			switch {
			case '0' <= s[i] && s[i] <= '9':
				escapedChar = escapedChar<<4 | (s[i] - '0')
				inEscape++
			case 'a' <= s[i] && s[i] <= 'f':
				escapedChar = escapedChar<<4 | (s[i] - 'a' + 10)
				inEscape++
			case 'A' <= s[i] && s[i] <= 'F':
				escapedChar = escapedChar<<4 | (s[i] - 'A' + 10)
				inEscape++
			default:
				next = s[i]
				inEscape = 0
				escapedChar = 0
			}
			if inEscape == 3 {
				next = escapedChar
				escapedChar = 0
				inEscape = 0
			}
		}

		switch next {
		case 0:
		case '"':
			ret += `""`
		default:
			ret += string(next)
		}
	}
	return
}
