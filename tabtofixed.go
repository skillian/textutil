package textutil

import (
	"fmt"
	"math/bits"
	"strings"
	"unicode"
)

type TabbedConfig struct {
	TabSize int
	Props map[string]any
}

func NewTabFixer(tabLen int) func(string) (string, error) {
	return newGridTabFixer(TabbedConfig{TabSize: tabLen}, tabbedGridConfig{})
}

func SQLInsert(tc TabbedConfig) func(string) (string, error) {
	tgc := tabbedGridConfig{
		GridFuncs: []func(tc TabbedConfig, grid [][]string) ([][]string, error){
			func(tc TabbedConfig, grid [][]string) ([][]string, error) {
				if len(grid) < 2 {
					return grid, nil
				}
				for i, header := range grid[0] {
					grid[0][i] = strings.Join([]string{"\"", "\","}, header)
				}
				// remove trailing comma from last header:
				lastHeader := &grid[0][len(grid[0])-1]
				*lastHeader = (*lastHeader)[:len(*lastHeader)-1]
				grid[0] = append(grid[0], ") VALUES (")
				const skipRows = 1
				for i, row := range grid[skipRows:] {
					for j, field := range row {
						switch {
						case strings.EqualFold(field, "null") || strings.HasPrefix(field, "@"):
							// pass
						default:
							field = strings.Join(strings.Split(field, "'"), "''")
							field = strings.Join([]string{"'", "'"}, field)
						}
						row[j] = field + ","
					}
					// remove trailing comma from last field:
					lastField := &row[len(row)-1]
					*lastField = (*lastField)[:len(*lastField)-1]
					grid[skipRows+i] = append(row, "), (")
				}
				lastRow := grid[len(grid)-1]
				lastRow[len(lastRow)-1] = ");"
				return grid, nil
			},
		},
		PostFuncs: []func(TabbedConfig, string) (string, error){
			func(tc TabbedConfig, s string) (string, error) {
				lines := strings.Split(s, EndLine)
				if len(lines) < 2 {
					return s, nil
				}
				for i, line := range lines {
					lines[i] = "\t" + strings.TrimRightFunc(line, unicode.IsSpace)
				}
				tableName := "tableName"
				if tn, ok := tc.Props["tablename"]; ok {
					tableName = fmt.Sprint(tn)
				}
				return strings.Join([]string{
					"INSERT INTO \"",
					tableName,
					"\" (",
					EndLine,
					lines[0],
					EndLine,
					strings.Join(lines[1:], EndLine),
				}, ""), nil
			},
		},
	}
	return newGridTabFixer(tc, tgc)
}

type tabbedGridConfig struct {
	GridFuncs []func(TabbedConfig, [][]string) ([][]string, error)
	PostFuncs []func(TabbedConfig, string) (string, error)
}

func newGridTabFixer(tc TabbedConfig, cfg tabbedGridConfig) func(string) (string, error) {
	return func(text string) (string, error) {
		grid, maxColLengths := splitGrid(text, EndLine, "\t")
		for _, gf := range cfg.GridFuncs {
			var err error
			grid, err = gf(tc, grid)
			if err != nil {
				return "", err
			}
		}
		if len(cfg.GridFuncs) > 0 && len(grid) > 0 {
			if len(grid[0]) != len(maxColLengths) {
				maxColLengths = make([]int, len(grid[0]))
			}
			for _, fields := range grid {
				for i, field := range fields {
					if len(field) > maxColLengths[i] {
						maxColLengths[i] = len(field)
					}
				}
			}
		}
		colPaddings := make([]string, len(maxColLengths))
		for i, length := range maxColLengths {
			length = (length / tc.TabSize) + 1
			colPaddings[i] = strings.Repeat("\t", length)
			maxColLengths[i] = length * tc.TabSize
		}
		sb := make([]byte, 0, 1<<bits.Len(uint(len(text)+1)))
		for _, fields := range grid {
			for i, field := range fields {
				paddingSpacesCount := maxColLengths[i] - len(field)
				paddingTabsCount := paddingSpacesCount / tc.TabSize
				if paddingTabsCount*tc.TabSize < paddingSpacesCount {
					paddingTabsCount++
				}
				sb = append(sb, []byte(field)...)
				sb = append(sb, []byte(colPaddings[i][:paddingTabsCount])...)
			}
			sb = append(sb, []byte(EndLine)...)
		}
		s := string(sb)
		for _, pf := range cfg.PostFuncs {
			var err error
			s, err = pf(tc, s)
			if err != nil {
				return "", err
			}
		}
		return s, nil
	}
}

func splitGrid(text, lineSep, fieldSep string) (grid [][]string, maxColLengths []int) {
	lines := strings.Split(text, lineSep)
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			break
		}
		lines = lines[:i]
	}
	grid = make([][]string, len(lines))
	fields := strings.Split(lines[0], fieldSep)
	maxColLengths = make([]int, len(fields))
	for i, field := range fields {
		maxColLengths[i] = len(field)
	}
	grid[0] = fields
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		fields = strings.Split(line, fieldSep)
		grid[i] = fields
		for i, field := range fields {
			if maxColLengths[i] < len(field) {
				maxColLengths[i] = len(field)
			}
		}
	}
	return
}
