package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	clipboardpkg "github.com/atotto/clipboard"
	"github.com/skillian/argparse"
	"github.com/skillian/errors"
	"github.com/skillian/textutil"
)

const (
	defaultFieldSep  = "\t"
	defaultRecordSep = textutil.EndLine
)

func main() {
	var isCsv bool
	var input, output string
	var tabSize int
	var transform interface{}
	var fieldSep, recordSep string
	parser := argparse.MustNewArgumentParser(
		argparse.Description(
			"Utility for transforming text",
		),
	)
	parser.MustAddArgument(
		argparse.OptionStrings("-c", "--csv"),
		argparse.ActionFunc(argparse.StoreTrue),
		argparse.Default(false),
		argparse.Help("Specify source is a CSV"),
	).MustBind(&isCsv)
	parser.MustAddArgument(
		argparse.OptionStrings("-f", "--field"),
		argparse.ActionFunc(argparse.Store),
		argparse.Default(defaultFieldSep),
		argparse.Help(
			"Specify a custom field separator in the text (default: %q)",
			defaultFieldSep,
		),
	).MustBind(&fieldSep)
	parser.MustAddArgument(
		argparse.OptionStrings("-r", "--record"),
		argparse.ActionFunc(argparse.Store),
		argparse.Default(defaultRecordSep),
		argparse.Help(
			"Specify a custom record separator in the text (default: %q)",
			defaultRecordSep,
		),
	).MustBind(&recordSep)
	parser.MustAddArgument(
		argparse.OptionStrings("-i", "--input"),
		argparse.ActionFunc(argparse.Store),
		argparse.Default(""),
		argparse.Help(
			"optional input file to use instead of the clipboard",
		),
	).MustBind(&input)
	parser.MustAddArgument(
		argparse.OptionStrings("-o", "--output"),
		argparse.ActionFunc(argparse.Store),
		argparse.Default(""),
		argparse.Help(
			"optional output file to use instead of the clipboard",
		),
	).MustBind(&output)
	const defaultTabSize = 8
	parser.MustAddArgument(
		argparse.OptionStrings("--tab-size"),
		argparse.ActionFunc(argparse.Store),
		argparse.Type(argparse.Int),
		argparse.Default(defaultTabSize),
		argparse.Help(
			"Specify the size of tabs (default: %d)",
			defaultTabSize,
		),
	).MustBind(&tabSize)
	type griddererFunc func(s string) (textutil.Gridder, error)
	var gridderer griddererFunc
	parser.MustAddArgument(
		argparse.OptionStrings("-s", "--source"),
		argparse.ActionFunc(argparse.Store),
		argparse.Default("tsv"),
		argparse.Help(
			"source format (default: tab-separated \"tsv\")",
		),
		argparse.Choices(
			argparse.Choice{
				Key: "csv",
				Value: griddererFunc(func(s string) (textutil.Gridder, error) {
					return textutil.CsvGridder{
						Reader: strings.NewReader(s),
					}, nil
				}),
				Help: "Comma-separated values",
			},
			argparse.Choice{
				Key: "tsv",
				Value: griddererFunc(func(s string) (textutil.Gridder, error) {
					return textutil.TextSplitGridder{
						Text:     s,
						LineSep:  "\n",
						FieldSep: "\t",
					}, nil
				}),
				Help: "Unquoted tab and newline-separated values and records",
			},
			argparse.Choice{
				Key: "chamentityxml",
				Value: griddererFunc(func(s string) (textutil.Gridder, error) {
					return textutil.ChamaeleonEntityXMLGridder{
						XMLText: s,
					}, nil
				}),
				Help: "Chamaeleon XML Entity array",
			},
		),
	).MustBind(&gridderer)
	parser.MustAddArgument(
		argparse.OptionStrings("-t", "--transform"),
		argparse.ActionFunc(argparse.Store),
		argparse.Help(
			"transform operation to run on the text",
		),
		argparse.Choices(
			argparse.Choice{
				Key:   "htmltable",
				Value: textutil.HTMLTable,
				Help:  "convert tsv into HTML table",
			},
			argparse.Choice{
				Key:   "jsonescape",
				Value: textutil.JSONEscape,
				Help:  "escape text into a JSON string",
			},
			argparse.Choice{
				Key:   "jsonunescape",
				Value: textutil.JSONUnescape,
				Help:  "unescape JSON into a string",
			},
			argparse.Choice{
				Key:   "sqlinsert",
				Value: textutil.SQLInsert,
				Help:  "translate a tab-delimited table with column headers into a SQL insert",
			},
			argparse.Choice{
				Key:   "sqlescape",
				Value: textutil.SQLEscape,
				Help:  "escape text so into a SQL string literal",
			},
			argparse.Choice{
				Key:   "sqlunescape",
				Value: textutil.SQLUnescape,
				Help:  "'unescape' a SQL string literal into a 'raw' value",
			},
			argparse.Choice{
				Key:   "tabtofixed",
				Value: textutil.NewTabFixer,
				Help:  "add extra tabs to tab-separated values so that columns line up",
			},
		),
	).MustBind(&transform)
	propsArg := parser.MustAddArgument(
		argparse.OptionStrings("-p", "--property"),
		argparse.Nargs(2),
		argparse.MetaVar("NAME", "VALUE"),
		argparse.ActionFunc(argparse.Append),
		argparse.Help("property keys and values used when formatting"),
	)
	ns := parser.MustParseArgs()
	var reader readAller
	if input != "" {
		reader = filename(input)
	} else {
		reader = clipboard{}
	}
	var writer writeAller
	if output != "" {
		writer = filename(output)
	} else {
		writer = clipboard{}
	}
	text, err := reader.readAll()
	if err != nil {
		panic(err)
	}
	tc := textutil.TabbedConfig{
		TabSize: tabSize,
	}
	if v, ok := ns.Get(propsArg); ok {
		vs := v.([]any)
		tc.Props = make(map[string]any, len(vs))
		for _, v := range vs {
			w := v.([]any)
			tc.Props[fmt.Sprint(w[0])] = w[1]
		}
	}
	var gridder textutil.Gridder
	switch f := transform.(type) {
	case func(textutil.TabbedConfig, string) (string, error):
		text, err = f(tc, text)
	case func(textutil.TabbedConfig) func(string) (string, error):
		text, err = f(tc)(text)
	case func(textutil.TabbedConfig, textutil.Gridder) (string, error):
		gridder, err = gridderer(text)
		if err == nil {
			text, err = f(tc, gridder)
		}
	case func(textutil.TabbedConfig) func(textutil.Gridder) (string, error):
		gridder, err = gridderer(text)
		if err == nil {
			text, err = f(tc)(gridder)
		}
	default:
		panic(errors.Errorf("unknown transformer %[1]v (type: %[1]T)", f))
	}
	if err != nil {
		panic(err)
	}
	if err = writer.writeAll(text); err != nil {
		panic(err)
	}
}

type readAller interface {
	readAll() (string, error)
}

type writeAller interface {
	writeAll(string) error
}

type clipboard struct{}

func (clipboard) readAll() (string, error) { return clipboardpkg.ReadAll() }
func (clipboard) writeAll(s string) error  { return clipboardpkg.WriteAll(s) }

type filename string

func (f filename) readAll() (string, error) {
	var bs []byte
	var err error
	if f == "" || f == "-" {
		bs, err = ioutil.ReadAll(os.Stdin)
	} else {
		bs, err = ioutil.ReadFile(string(f))
	}
	return string(bs), err
}

func (f filename) writeAll(s string) error {
	if f == "" || f == "-" {
		_, err := io.Copy(os.Stdout, strings.NewReader(s))
		return err
	}
	return ioutil.WriteFile(string(f), []byte(s), 0)
}
