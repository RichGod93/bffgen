package commands

import (
	"fmt"
	"os"

	"github.com/RichGod93/bffgen/internal/utils"
	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert configuration between YAML and JSON formats",
	Long:  `Convert bff.config.yaml to bffgen.config.json or vice versa.`,
}

var (
	fromFormat string
	toFormat   string
	outputFile string
)

var convertConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Convert config file format",
	Long:  `Convert between bff.config.yaml (Go) and bffgen.config.json (Node.js) formats.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := convertConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Conversion failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Configuration converted successfully!")
	},
}

func init() {
	convertConfigCmd.Flags().StringVar(&fromFormat, "from", "", "Source format (yaml or json)")
	convertConfigCmd.Flags().StringVar(&toFormat, "to", "", "Target format (yaml or json)")
	convertConfigCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: auto-detected)")

	convertCmd.AddCommand(convertConfigCmd)
	rootCmd.AddCommand(convertCmd)
}

func convertConfig() error {
	// Auto-detect formats if not specified
	if fromFormat == "" {
		if _, err := os.Stat("bff.config.yaml"); err == nil {
			fromFormat = "yaml"
		} else if _, err := os.Stat("bffgen.config.json"); err == nil {
			fromFormat = "json"
		} else {
			return fmt.Errorf("no config file found (bff.config.yaml or bffgen.config.json)")
		}
	}

	if toFormat == "" {
		if fromFormat == "yaml" {
			toFormat = "json"
		} else {
			toFormat = "yaml"
		}
	}

	// Validate formats
	if fromFormat != "yaml" && fromFormat != "json" {
		return fmt.Errorf("invalid source format: %s (must be yaml or json)", fromFormat)
	}
	if toFormat != "yaml" && toFormat != "json" {
		return fmt.Errorf("invalid target format: %s (must be yaml or json)", toFormat)
	}

	if fromFormat == toFormat {
		return fmt.Errorf("source and target formats are the same")
	}

	fmt.Printf("🔄 Converting from %s to %s\n", fromFormat, toFormat)

	// Perform conversion
	if fromFormat == "yaml" && toFormat == "json" {
		return utils.ConvertYAMLToJSON(outputFile)
	}

	return utils.ConvertJSONToYAML(outputFile)
}
