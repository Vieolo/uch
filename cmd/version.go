package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"gopkg.in/yaml.v3"
)

// The bytes is injected from main.go downward
var ThisGyByte []byte

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the version of uch",
	Long:  "Displays the version of uch",
	Run: func(cmd *cobra.Command, args []string) {
		type gyStruct struct {
			Version string `yaml:"version"`
		}

		var gy gyStruct
		err := yaml.Unmarshal(ThisGyByte, &gy)
		if err != nil {
			termange.PrintErrorln(err.Error())
			return
		}
		termange.PrintInfof("v%s\n", gy.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
