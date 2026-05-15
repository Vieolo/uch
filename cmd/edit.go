package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/vieolo/termange"
	"github.com/vieolo/uch/internal/utils"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the central command storage JSON",
	Long:  `Opens the central command storage JSON in terminal, allowing you to edit and save it`,
	Run: func(cmd *cobra.Command, args []string) {
		tempFile, err := os.CreateTemp("", "uch-temp-*.json")
		if err != nil {
			termange.PrintErrorf("There was an error creating the temporary file: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(tempFile.Name())

		centCom, err := utils.GetCentCom()
		if err != nil {
			termange.PrintErrorf("There was an error getting the config.json: %v\n", err)
			os.Exit(1)
		}

		marsh, err := json.MarshalIndent(centCom, "", "  ")
		if err != nil {
			termange.PrintErrorf("There was an error loading config.json into memory: %v\n", err)
			os.Exit(1)
		}
		tempFile.Write(marsh)
		tempFile.Close()

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim" // Default fallback
		}

		// Prepare the command
		c := exec.Command(editor, tempFile.Name())

		// Hook up the OS streams to the command
		// This allows Vim to take over the terminal screen and receive keyboard input
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		// Run the command and wait for the user to exit Vim
		fmt.Printf("Opening config.json in %s...\n", editor)
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running editor: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("File successfully edited and saved!")

		// Optional: Read the file back into your program if you need to process it further
		/*
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}
			fmt.Printf("New content:\n%s\n", string(data))
		*/
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
