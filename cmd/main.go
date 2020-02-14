package main

import (
	"bufio"
	"fmt"
	"log"
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
	arrayUsage         = "Make array of root objects"

	filesFlagShorthand = "f"
	filesFlag          = "files"
	filesUsage         = "Bind files"

	noSortKeysFlagShorthand = "n"
	noSortKeysFlag          = "nosort"
	noSortKeysUsage         = "Do not sort keys in objects"

	schemaFlagShorthand = "s"
	schemaFlag          = "schema"
	schemaUsage         = "Path to YAML schema"
)

const fileFlagPrefix = "f"

func init() {
	log.SetPrefix("")
}

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	schemaPath := fs.StringP(schemaFlag, schemaFlagShorthand, "", schemaUsage)
	files := fs.StringToStringP(filesFlag, filesFlagShorthand, map[string]string{}, filesUsage)
	noSortKeys := fs.BoolP(noSortKeysFlag, noSortKeysFlagShorthand, false, noSortKeysUsage)
	var arrayLen schema.Length
	fs.VarP(&arrayLen, arrayFlag, arrayFlagShorthand, arrayUsage)

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if flag.NArg() > 0 {
		return fmt.Errorf("no additional args expected, got: %s", flag.Args())
	}
	if *schemaPath == "" {
		return fmt.Errorf("no schema provided")
	}

	f, err := os.Open(*schemaPath)
	if err != nil {
		return fmt.Errorf("unable to open schema %q: %w", schemaPath, err)
	}

	var sch schema.Schema
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&sch); err != nil {
		return fmt.Errorf("unable to unmarshal schema: %w", err)
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
		if err := ctx.AddFile(name, &schema.File{Path: file}); err != nil {
			return fmt.Errorf("unable to add file %q: %w", name, err)
		}
	}

	rand.Seed(time.Now().UnixNano())
	bw := bufio.NewWriterSize(os.Stdout, 1024)
	defer bw.Flush()

	if err := sch.GenerateJSON(ctx, bw, &arrayLen); err != nil {
		return fmt.Errorf("unable to generate: %w", err)
	}

	return nil
}
