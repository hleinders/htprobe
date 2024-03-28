/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"strings"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type ContentFlags struct {
	follow bool
}

var contentFlags ContentFlags

var contentShortDesc = "Makes a http request and displays the content of the response, if any"

// contentCmd represents the content command
var contentCmd = &cobra.Command{
	Use:     "content <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"cnt", "cont"},
	Short:   contentShortDesc,
	Long: makeHeader("htprobe content: "+contentShortDesc) + `With 'htprobe content <URL>', the full response body
	is shown. You may pass the '-f|--follow' flag to follow redirects.
	In this case, the content of any hop is displayed with the '-a|--all' flag.

	Flags marked with '***' may be used multiple times.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecContent(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(contentCmd)

	// flags
	contentCmd.Flags().BoolVarP(&contentFlags.follow, "follow", "f", false, "show all response cookies")

}

func ExecContent(cmd *cobra.Command, args []string) {
	var hops []WebRequestResult
	var err error

	// create template request:
	req := WebRequest{
		agent:     rootFlags.agent,
		lang:      rootFlags.reqLang,
		method:    rootFlags.httpMethod,
		authUser:  rootFlags.authUser,
		authPass:  rootFlags.authPass,
		reqBody:   globalRequestBody,
		xhdrs:     globalHeaderList,
		cookieLst: globalCookieLst,
	}

	for _, rawURL := range args {
		newReq := req
		newReq.url, err = checkURL(rawURL)
		check(err, ErrNoURL)

		// handle the request(s)
		if headerFlags.follow {
			hops, err = follow(&newReq, &connSet)
			if err != nil {
				pr.Error(err.Error())
			}
		} else {
			hops, err = noFollow(&newReq, &connSet)
			if err != nil {
				pr.Error(err.Error())
			}

		}

		// display results
		prettyPrintContent(hops)
	}
}

func prettyPrintContent(resultList []WebRequestResult) {

	fmt.Println()

	for cnt, h := range resultList {

		title := fmt.Sprintf("%d:  %s (%s)", cnt+1, h.PrettyPrintRedir(cnt), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Println(title)
		fmt.Println(strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Println()

		body, err := io.ReadAll(h.response.Body)
		if err != nil {
			pr.Errorln("%s", err)
		} else {
			fmt.Println(at.Bold("Content:"))
			fmt.Println(at.Bold(strings.Repeat(at.FrameOHLine, 8)))
			fmt.Printf("\n%+v\n", string(body))
		}

		fmt.Println()
	}
}
