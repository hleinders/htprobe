/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionShortDesc = "Shows version and runtime information"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"vers"},
	Short:   versionShortDesc,
	Long:    makeHeader(lowerAppName+" version: "+versionShortDesc) + `This subcommand does not take any arguments or options.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecVersions(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func ExecVersions(cmd *cobra.Command, args []string) {
	fmt.Println("")
	fmt.Printf("Version:     %s\n", AppVersion)
	fmt.Printf("Go version:  %s\n", runtime.Version())
	fmt.Printf("Go compiler: %s\n", runtime.Compiler)
	fmt.Printf("Binary type: %s (%s)\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Author:      %s\n", Author)
}
