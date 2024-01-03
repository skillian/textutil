package textutil

import (
	"fmt"
	"strings"
)

func SQLEscape(tabLen int) func(string) (string, error) {
	return func(text string) (string, error) {
		sb := strings.Builder{}
		sb.WriteRune('\'')
		for _, r := range text {
			switch r {
			case '\'':
				sb.WriteString("''")
			default:
				sb.WriteRune(r)
			}
		}
		sb.WriteRune('\'')
		return sb.String(), nil
	}
}

func SQLUnescape(tabLen int) func(string) (string, error) {
	return func(text string) (string, error) {
		text = strings.TrimSuffix(strings.TrimPrefix(text, "'"), "'")
		{
			// avoid an allocation if we don't need it:
			hasQuote := false
			for _, c := range []byte(text) {
				if c == '\'' {
					hasQuote = true
					break
				}
			}
			if !hasQuote {
				return text, nil
			}
		}
		sb := strings.Builder{}
		lastIsQuote := false
		for i, r := range text {
			switch r {
			case '\'':
				if !lastIsQuote {
					lastIsQuote = true
					continue
				}
				fallthrough
			default:
				if lastIsQuote && r != '\'' {
					return "", fmt.Errorf(
						"SQLUnscape:%d: Single quote not followed by another",
						i+1,
					)
				}
				sb.WriteRune(r)
				lastIsQuote = false
			}
		}
		return sb.String(), nil
	}
}
