package main

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	clipboardpkg "github.com/atotto/clipboard"
	"github.com/skillian/argparse"
	"github.com/skillian/errors"
	"github.com/skillian/textutil"
)

func main() {
	var input, output string
	var tabSize int
	var transform interface{}
	parser := argparse.MustNewArgumentParser(
		argparse.Description(
			"Utility for transforming text",
		),
	)
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
	parser.MustAddArgument(
		argparse.OptionStrings("--tab-size"),
		argparse.ActionFunc(argparse.Store),
		argparse.Type(argparse.Int),
		argparse.Default(8),
		argparse.Help(
			"Specify the size of tabs (default: 8)",
		),
	).MustBind(&tabSize)
	parser.MustAddArgument(
		argparse.OptionStrings("-t", "--transform"),
		argparse.ActionFunc(argparse.Store),
		argparse.Help(
			"transform operation to run on the text",
		),
		argparse.Choices(
			argparse.Choice{
				Key:   "jsonescape",
				Value: textutil.JSONEscape,
				Help:  "escape text into a JSON string",
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
	_ = parser.MustParseArgs()
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
	transformerFunc, ok := transform.(func(string) (string, error))
	if !ok {
		switch x := transform.(type) {
		case func(int) func(string) (string, error):
			transformerFunc = x(tabSize)
		default:
			panic(errors.Errorf("unknown transformer %[1]v (type: %[1]T)", x))
		}
	}
	if text, err = transformerFunc(text); err != nil {
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
