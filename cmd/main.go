package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mitinarseny/jg/schema"
	"gopkg.in/yaml.v3"
)

const (
	arrayFlag = "array"
)

func main() {
	arrayLen := flag.Int(arrayFlag, -1, "array [N]")
	flag.Parse()

	if nArgs := flag.NArg(); nArgs != 1 {
		log.Printf("expected only 1 arg, got: %d\n", nArgs)
		os.Exit(1)
	}

	if err := run(flag.Arg(0), *arrayLen); err != nil {
		log.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func run(schemaPath string, arrayLen int) error {
	f, err := os.Open(schemaPath)
	if err != nil {
		return fmt.Errorf("unable to open %q: %w", schemaPath, err)
	}

	var sch schema.Schema
	if err := yaml.NewDecoder(f).Decode(&sch); err != nil {
		return fmt.Errorf("unable to unmarshal: %w", err)
	}

	if err := sch.Generate(os.Stdout, arrayLen); err != nil {
		return fmt.Errorf("unable to generate: %w", err)
	}

	return nil
}
