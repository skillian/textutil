package textutil

import (
	"encoding/json"
	"fmt"
	"strings"
)

func JSONUnescape(_ int) func(string) (string, error) {
	return func(s string) (string, error) {
		var v interface{}
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			if len(s) > 20 {
				s = s[:17] + "..."
			}
			return "", fmt.Errorf(
				"failed to parse %v as JSON: %w",
				s, err,
			)
		}
		return fmt.Sprint(v), nil
	}
}

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
