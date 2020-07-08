package main

import (
	"bufio"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	// _ "net/http/pprof"
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
	arrayUsage         = "Generate array of root objects (0 means do not wrap in array)"

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

	streamFlagShorthand = "s"
	streamFlag          = "stream"
	streamUsage         = "Stream root objects delimited by newline (-1 means endless)"
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
	stream := fs.Int64P(streamFlag, streamFlagShorthand, 0, streamUsage)
	var arrayLen schema.Length
	fs.VarP(&arrayLen, arrayFlag, arrayFlagShorthand, arrayUsage)

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

	if *stream != 0 && arrayLen.Max != 0 {
		fs.Usage()
		return fmt.Errorf("'--array' and '--stream' can not be used at the same time")
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

	var seed uint64
	if err := binary.Read(crand.Reader, binary.BigEndian, &seed); err != nil {
		seed = uint64(time.Now().UnixNano())
	}
	rnd := rand.New(rand.NewSource(int64(seed)))

	switch {
	case arrayLen.Max != 0:
		a := schema.Array{
			Length:   arrayLen,
			Elements: sch.Root,
		}
		return a.GenerateJSON(ctx, w, rnd)
	case *stream != 0:
		return sch.StreamJSON(ctx, w, rnd, *stream)
	default:
		return sch.GenerateJSON(ctx, w, rnd)
	}
}
