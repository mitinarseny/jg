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
	usageTemplate = `Usage: %s [OPTIONS] SCHEMA

JSON generator

Options:
%s
`

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
	outBuffSizeUsage   = "Buffer size for JSON output (0 means no buffer)"
	outBuffSizeDefault = 1024
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ff, err := os.Create("cpu.pprof")
	// if err != nil {
	// 	log.Fatal("could not create CPU profile: ", err)
	// }
	// defer ff.Close()
	// if err := pprof.StartCPUProfile(ff); err != nil {
	// 	log.Fatal("could not start CPU profile: ", err)
	// }
	// defer pprof.StopCPUProfile()


	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	files := fs.StringToStringP(filesFlag, filesFlagShorthand, map[string]string{}, filesUsage)
	noSortKeys := fs.BoolP(noSortKeysFlag, noSortKeysFlagShorthand, noSortKeysDefault, noSortKeysUsage)
	out := fs.StringP(outFlag, outFlagShorthand, outDefault, outUsage)
	outBuffSize := fs.Uint(outBuffSizeFlag, outBuffSizeDefault, outBuffSizeUsage)
	arrayLen := &schema.Length{}
	fs.VarP(arrayLen, arrayFlag, arrayFlagShorthand, arrayUsage)

	fs.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usageTemplate, os.Args[0], fs.FlagUsages())
	}

	switch err := fs.Parse(os.Args[1:]); err {
	case flag.ErrHelp:
		return nil
	default:
		return err
	case nil:
	}

	switch n := fs.NArg(); n {
	case 0:
		fs.Usage()
		return fmt.Errorf("no schema provided")
	case 1: // schemaPath
	default:
		fs.Usage()
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
		outFile, err = os.Create(*out)
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
	 // TODO: crypto/rand seed
	if err := sch.GenerateJSON(ctx, w, rand.New(rand.NewSource(time.Now().UnixNano())), arrayLen); err != nil {
		return fmt.Errorf("unable to generate: %w", err)
	}

	return nil
}
