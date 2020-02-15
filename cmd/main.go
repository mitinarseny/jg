package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/mitinarseny/jg/schema"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Flags
const (
	arrayFlagShorthand = "a"
	arrayFlag          = "array"
	arrayUsage         = "Generate array of root objects"

	filesFlagShorthand = "f"
	filesFlag          = "files"
	filesUsage         = "Bind files to their names in schema"

	noSortKeysFlagShorthand = "n"
	noSortKeysFlag          = "nosort"
	noSortKeysUsage         = "Do not sort keys in objects"
	noSortKeysDefault       = false

	outFlagShorthand = "o"
	outFlag          = "output"
	outUsage         = "JSON output"
	outDefault       = "/dev/stdout"

	outBuffSizeFlag    = "output-buff-size"
	outBuffSizeUsage   = "Buffer size for JSON output (0 or less means no buffer)"
	outBuffSizeDefault = 1024

	schemaFlagShorthand = "s"
	schemaFlag          = "schema"
	schemaUsage         = "Path to YAML schema"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	files := fs.StringToStringP(filesFlag, filesFlagShorthand, map[string]string{}, filesUsage)
	noSortKeys := fs.BoolP(noSortKeysFlag, noSortKeysFlagShorthand, noSortKeysDefault, noSortKeysUsage)
	out := fs.StringP(outFlag, outFlagShorthand, outDefault, outUsage)
	outBuffSize := fs.Uint(outBuffSizeFlag, outBuffSizeDefault, outBuffSizeUsage)
	arrayLen := &schema.Length{}
	fs.VarP(arrayLen, arrayFlag, arrayFlagShorthand, arrayUsage)

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	switch n := fs.NArg(); n {
	case 1: // schemaPath
	case 0:
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n%s\n", os.Args[0], fs.FlagUsages())
		return fmt.Errorf("no schema provided")
	default:
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n%s\n", os.Args[0], fs.FlagUsages())
		return fmt.Errorf("only 1 positional arg expected, got: %d", n)
	}
	schemaPath := fs.Arg(0)

	f, err := os.Open(schemaPath)
	if err != nil {
		return err
	}

	var outFile *os.File
	if *out == "/dev/stdout" {
		outFile = os.Stdout
	} else {
		var err error
		outFile, err = os.Open(*out)
		if err != nil {
			return err
		}
	}

	var sch schema.Schema
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&sch); err != nil {
		return fmt.Errorf("unable to unmarshal schema %q: %w", schemaPath, err)
	}

	if err := sch.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	ctx := schema.NewContext()
	defer ctx.Close()
	ctx.SetSortKeys(!*noSortKeys)

	for name := range sch.Files {
		file, found := (*files)[name]
		if !found {
			return fmt.Errorf("file %q is not provided", name)
		}
		if err := ctx.AddFile(name, file); err != nil {
			return fmt.Errorf("unable to add file %q: %w", name, err)
		}
	}

	w := io.Writer(outFile)
	if *outBuffSize > 0 {
		bw := bufio.NewWriterSize(w, int(*outBuffSize))
		defer bw.Flush()
		w = bw
	}

	if arrayLen.Max == 0 {
		arrayLen = nil
	}
	rand.Seed(time.Now().UnixNano())
	if err := sch.GenerateJSON(ctx, w, arrayLen); err != nil {
		return fmt.Errorf("unable to generate: %w", err)
	}

	return nil
}
