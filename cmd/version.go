package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"gopkg.in/yaml.v3"
)

// ThisGyByte holds the bytes of the project's go.yaml, embedded by main.go.
// It is the single source of truth for the CLI version.
var ThisGyByte []byte

// currentVersion returns the version from the embedded go.yaml, or "dev" if
// the bytes are missing or unparseable (e.g. during tests).
func currentVersion() string {
	if len(ThisGyByte) == 0 {
		return "dev"
	}
	type gyStruct struct {
		Version string `yaml:"version"`
	}
	var gy gyStruct
	if err := yaml.Unmarshal(ThisGyByte, &gy); err != nil || gy.Version == "" {
		return "dev"
	}
	return gy.Version
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of uch",
	Long:  "Displays the version of uch",
	Run: func(cmd *cobra.Command, args []string) {
		termange.PrintInfof("v%s\n", currentVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
