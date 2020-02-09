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
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.SortFlags = false

	schemaPath := fs.StringP(schemaFlag, schemaFlagShorthand, "", schemaUsage)
	noSortKeys := fs.BoolP(noSortKeysFlag, noSortKeysFlagShorthand, false, noSortKeysUsage)
	var arrayLen schema.Length
	fs.VarP(&arrayLen, arrayFlag, arrayFlagShorthand, arrayUsage)

	fs.ParseErrorsWhitelist.UnknownFlags = true // this is the reason why not standard flag package
	_ = fs.Parse(os.Args[1:])

	if flag.NArg() > 0 {
		return fmt.Errorf("no additional args expected, got: %s", flag.Args())
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

	files, filesFS := makeFileFlags(&sch)

	fs.AddFlagSet(filesFS)
	fs.ParseErrorsWhitelist.UnknownFlags = true // now we have defined all flags
	_ = fs.Parse(os.Args[1:])

	ctx := schema.NewContext()
	ctx.SetSortKeys(!*noSortKeys)

	for name, file := range files {
		if file.Path == "" {
			return fmt.Errorf("file %q not provided", name)
		}
		if err := ctx.AddFile(name, file); err != nil {
			return fmt.Errorf("unable to add files %q: %w", name, err)
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

func makeFileFlags(sch *schema.Schema) (map[string]*schema.File, *flag.FlagSet) {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	files := make(map[string]*schema.File, len(sch.Files))
	for name := range sch.Files {
		var f schema.File
		fs.VarP(&f, fileFlagPrefix+name, "", fmt.Sprintf("Data for %q", name))
		files[name] = &f
	}
	return files, fs
}
