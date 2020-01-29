package cmd

import (
	"fmt"
	"log"
	"os"

	schema "github.com/mitinarseny/jg/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(args[0]); err != nil {
			log.Print(err)
			os.Exit(1)
		}
	},
}

func run(schemaPath string) error {
	f, err := os.Open(schemaPath)
	if err != nil {
		return fmt.Errorf("unable to open %q: %w", schemaPath, err)
	}

	var sch schema.Schema
	if err := yaml.NewDecoder(f).Decode(&sch); err != nil {
		return fmt.Errorf("unable to unmarshal: %w", err)
	}

	if err := sch.Generate(os.Stdout); err != nil {
		return fmt.Errorf("unable to generate: %w", err)
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
