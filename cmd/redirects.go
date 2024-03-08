/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type RedirectFlags struct {
	showResponseHeader, showRequestHeader bool
	allHops                               bool
	showContent, showCert                 bool
	acceptCookies, showCookies            bool
	unsorted, showCookiesDetails          bool
	addHeaders, displaySingleHeader       []string
	bodyFile                              string
	bodyValues                            []string
}

// redirectsCmd represents the redirects command
var redirectsCmd = &cobra.Command{
	Use:   "redirects <URL>",
	Args:  cobra.MinimumNArgs(1),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("redirects called")
		ExecRedirects(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(redirectsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// redirectsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// redirectsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func ExecRedirects(cmd *cobra.Command, args []string) {
	fmt.Println("This is:", cmd.Use)
	fmt.Println("Args:", args)
	fmt.Printf("RootFlags: %+v\n", rootFlags)
	fmt.Printf("ConnSet:  %+v\n", connSet)
}
