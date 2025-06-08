/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package android

import (
	"fmt"

	"github.com/spf13/cobra"
	tu "github.com/vieolo/terminal-utils"
)

// connectDeviceCmd represents the connectDevice command
var ConnectDeviceCmd = &cobra.Command{
	Use:   "connectDevice",
	Short: "Connects to the physical device in your network",
	Long:  `Connects to the physical device in your network`,
	Run: func(cmd *cobra.Command, args []string) {
		deviceIP := ""
		if len(args) == 1 {
			deviceIP = args[0]
		}

		if deviceIP == "" {
			deviceIP = tu.TextInput("What is the IP of the device?")
		}

		tu.RunRawCommand("adb tcpip 5555")
		tu.RunRawCommand(fmt.Sprintf("adb connect %v:5555", deviceIP))
	},
}

func init() {
	AndroidCmd.AddCommand(ConnectDeviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// connectDeviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// connectDeviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
