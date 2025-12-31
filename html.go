package textutil

import (
	"fmt"
	"html"
	"strings"
)

func HTMLTable(tc TabbedConfig) func(string) (string, error) {
	var paragraphStyle string
	if paragraphStyleAny, ok := tc.Props["pstyle"]; ok {
		paragraphStyle = fmt.Sprintf(`style="%s"`, paragraphStyleAny)
	}
	return func(s string) (string, error) {
		sb := strings.Builder{}
		for i, line := range strings.Split(strings.TrimSpace(s), "\n") {
			if i == 0 {
				sb.WriteString("<table><tbody>")
			}
			sb.WriteString("<tr>")
			for _, field := range strings.Split(line, "\t") {
				sb.WriteString("<td><p")
				sb.WriteString(paragraphStyle)
				sb.WriteString(">")
				sb.WriteString(html.EscapeString(field))
				sb.WriteString("</p></td>")
			}
			sb.WriteString("</tr>")
		}
		if sb.Len() > 0 {
			sb.WriteString("</tbody></table>")
		}
		return sb.String(), nil
	}
}
