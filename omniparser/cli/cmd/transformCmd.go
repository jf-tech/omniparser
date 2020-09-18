package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/omniparser"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
	"github.com/jf-tech/omniparser/strs"
)

var (
	transformCmd = &cobra.Command{
		Use:   "transform",
		Short: "Transforms input to desired output based on a schema.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return doTransform()
		},
	}
	schema string
	input  string
)

func init() {
	transformCmd.Flags().StringVarP(
		&schema, "schema", "s", "", "an omniparser schema file (required)")
	_ = transformCmd.MarkFlagRequired("schema")

	transformCmd.Flags().StringVarP(
		&input, "input", "i", "", "an input file to be transformed (optional; if not specified, stdin/pipe is used)")
}

func fileExists(file string) bool {
	fi, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return !fi.IsDir()
}

func processFileFlag(fileLabel string, file string) (io.ReadCloser, error) {
	if !fileExists(schema) {
		return nil, fmt.Errorf("%s file '%s' does not exist", fileLabel, file)
	}
	return os.Open(file)
}

func doTransform() error {
	schemaName := filepath.Base(schema)
	schemaReadCloser, err := processFileFlag("schema", schema)
	if err != nil {
		return err
	}
	defer schemaReadCloser.Close()

	inputReadCloser := io.ReadCloser(nil)
	inputName := ""
	if strs.IsStrNonBlank(input) {
		inputName = filepath.Base(input)
		inputReadCloser, err = processFileFlag("input", input)
		if err != nil {
			return err
		}
		defer inputReadCloser.Close()
	} else {
		inputName = "(stdin)"
		inputReadCloser = os.Stdin
		// Note we don't defer Close() on this since os/golang runtime owns it.
	}

	parser, err := omniparser.NewParser(schemaName, schemaReadCloser)
	if err != nil {
		return err
	}

	op, err := parser.GetTransformOp(inputName, inputReadCloser, &transformctx.Ctx{})
	if err != nil {
		return err
	}

	fmt.Println("[")
	printOne := func() error {
		b, err := op.Read()
		if err != nil {
			return err
		}
		fmt.Print(
			strings.Join(
				strs.NoErrMapSlice(
					strings.Split(jsons.BPJ(string(b)), "\n"),
					func(s string) string { return "\t" + s }),
				"\n"))
		return nil
	}
	if op.Next() {
		if err := printOne(); err != nil {
			return err
		}
		for op.Next() {
			fmt.Print(",\n")
			if err := printOne(); err != nil {
				return err
			}
		}
	}
	fmt.Println("\n]")
	return nil
}
