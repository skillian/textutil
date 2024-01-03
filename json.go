package textutil

import (
	"strings"
)

func JSONEscape(_ int) func(string) (string, error) {
	return func(text string) (string, error) {
		sb := strings.Builder{}
		sb.WriteByte('"')
		for _, r := range text {
			switch r {
			case '"', '\n', '\r', '\t':
				sb.WriteRune('\\')
				switch r {
				case '\n':
					r = 'n'
				case '\r':
					r = 'r'
				case '\t':
					r = 't'
				}
				fallthrough
			default:
				sb.WriteRune(r)
			}
		}
		sb.WriteRune('"')
		return sb.String(), nil
	}
}
