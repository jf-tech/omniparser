package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jf-tech/go-corelib/ios"
	"github.com/jf-tech/go-corelib/jsons"
	"github.com/jf-tech/go-corelib/strs"
	"github.com/spf13/cobra"

	"github.com/jf-tech/omniparser"
	"github.com/jf-tech/omniparser/transformctx"
)

var (
	transformCmd = &cobra.Command{
		Use:   "transform",
		Short: "Transforms input to desired output based on a schema.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := doTransform(); err != nil {
				fmt.Println() // to sure cobra cli always write out "Error: ..." on a new line.
				return err
			}
			return nil
		},
	}
	schema string
	input  string
	ndjson bool
)

func init() {
	transformCmd.Flags().StringVarP(&schema, "schema", "s", "", "schema file (required)")
	_ = transformCmd.MarkFlagRequired("schema")

	transformCmd.Flags().StringVarP(
		&input, "input", "i", "", "input file (optional; if not specified, stdin/pipe is used)")
	transformCmd.Flags().BoolVarP(
		&ndjson, "stream", "", false, "change the output format to ndjson")
}

func openFile(label string, filepath string) (io.ReadCloser, error) {
	if !ios.FileExists(schema) {
		return nil, fmt.Errorf("%s file '%s' does not exist", label, filepath)
	}
	return os.Open(filepath)
}

func doTransform() error {
	schemaName := filepath.Base(schema)
	schemaReadCloser, err := openFile("schema", schema)
	if err != nil {
		return err
	}
	defer schemaReadCloser.Close()

	inputReadCloser := io.ReadCloser(nil)
	inputName := ""
	if strs.IsStrNonBlank(input) {
		inputName = filepath.Base(input)
		inputReadCloser, err = openFile("input", input)
		if err != nil {
			return err
		}
		defer inputReadCloser.Close()
	} else {
		inputName = "(stdin)"
		inputReadCloser = os.Stdin
		// Note we don't defer Close() on this since os/golang runtime owns it.
	}

	schema, err := omniparser.NewSchema(schemaName, schemaReadCloser)
	if err != nil {
		return err
	}

	transform, err := schema.NewTransform(inputName, inputReadCloser, &transformctx.Ctx{})
	if err != nil {
		return err
	}

	doOne := func() (string, error) {
		b, err := transform.Read()
		if err != nil {
			return "", err
		}

		s := string(b)
		if ndjson {
			return s, nil
		}

		return strings.Join(
			strs.NoErrMapSlice(
				strings.Split(jsons.BPJ(s), "\n"),
				func(s string) string { return "\t" + s }),
			"\n"), nil
	}

	record, err := doOne()
	if err == io.EOF {
		if ndjson {
			fmt.Println("")
		} else {
			fmt.Println("[]")
		}
		return nil
	}
	if err != nil {
		return err
	}

	lparen := "[\n%s"
	delim := ",\n%s"
	rparen := "\n]"
	if ndjson {
		lparen = "%s"
		delim = "\n%s"
		rparen = ""
	}

	fmt.Printf(lparen, record)
	for {
		record, err = doOne()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf(delim, record)
	}
	fmt.Println(rparen)
	return nil
}
