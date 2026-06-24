package textutil

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"strings"
)

func min[T cmp.Ordered](first T, others ...T) T {
	for _, x := range others {
		if x < first {
			first = x
		}
	}
	return first
}

type ChamaeleonEntityXMLGridder struct {
	XMLText string
}

func (g ChamaeleonEntityXMLGridder) Grid() (grid [][]string, maxColLengths []int, err error) {
	var es chamXMLEntities
	if err = xml.Unmarshal([]byte(g.XMLText), &es); err != nil {
		return nil, nil, fmt.Errorf(
			"unmarshaling XML %q: %w",
			g.XMLText[:min(8, len(g.XMLText))],
			err,
		)
	}
	grid = make([][]string, len(es.Entities)+1)
	headers := make([]string, 0, 16)
	for i := range es.Entities {
		e := &es.Entities[i]
		for j := range e.Attrs {
			a := &e.Attrs[j]
			headerExists := false
			for _, h := range headers {
				headerExists = strings.EqualFold(a.Name, h)
				if headerExists {
					break
				}
			}
			if !headerExists {
				headers = append(headers, a.Name)
			}
		}
	}
	attrs := make([]string, len(es.Entities)*len(headers))
	grid[0] = headers
	for i := range es.Entities {
		e := &es.Entities[i]
		grid[i+1] = attrs[:len(headers)]
		attrs = attrs[len(headers):]
		for j, h := range headers {
			for k := range e.Attrs {
				a := &e.Attrs[k]
				if !strings.EqualFold(a.Name, h) {
					continue
				}
				grid[i+1][j] = a.Value
			}
		}
	}
	maxColLengths = getMaxColLengths(grid)
	return
}

type chamXMLEntities struct {
	Entities []chamXMLEntity `xml:"Entity"`
}

type chamXMLEntity struct {
	Name      string        `xml:",attr"`
	ID        string        `xml:"Id,attr"`
	IsDeleted bool          `xml:",attr"`
	IsDirty   bool          `xml:",attr"`
	Persisted bool          `xml:",attr"`
	Attrs     []chamXMLAttr `xml:"Attr"`
}

type chamXMLAttr struct {
	Name  string `xml:",attr"`
	Type  string `xml:",attr"`
	Value string `xml:",chardata"`
}
