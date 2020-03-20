package cluster

import "unicode"

func simpleName(input string) string {
	var last = '-'
	var result = make([]rune, 0, len(input))
	for _, r := range input {
		r = unicode.ToLower(r)
		switch {
		case 'a' <= r && r <= 'z':
			result = append(result, r)
			last = r
		case '0' <= r && r <= '9':
			result = append(result, r)
			last = r
		default:
			if last == '-' {
				continue
			}
			result = append(result, '-')
			last = '-'
		}
	}

	if len(result) == 0 {
		return ""
	}

	if last == '-' {
		return string(result[0 : len(result)-1])
	}
	return string(result)
}
