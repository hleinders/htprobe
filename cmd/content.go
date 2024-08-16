/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type ContentFlags struct {
	follow  bool
	outFile string
}

var contentFlags ContentFlags

var contentShortDesc = "Makes a http request and displays the content of the response, if any"

// contentCmd represents the content command
var contentCmd = &cobra.Command{
	Use:     "content <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"cnt", "cont"},
	Short:   contentShortDesc,
	Long: makeHeader(lowerAppName+" content: "+contentShortDesc) + `With command 'content', the full response body
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
	contentCmd.Flags().BoolVarP(&contentFlags.follow, "follow", "f", false, "show content for all hops")

	// Parameter
	contentCmd.Flags().StringVarP(&contentFlags.outFile, "outfile", "o", "", "write content to `file`")
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
		newReq.url, err = checkURL(rawURL, false)
		check(err, ErrNoURL)

		// handle the request(s)
		if contentFlags.follow {
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
	var fo *os.File
	var err error

	out := os.Stdout

	// only for screen
	fmt.Println()

	if contentFlags.outFile != "" {
		fo, err = os.Create(contentFlags.outFile)
		if err == nil {
			defer fo.Close()
			out = fo
		} else {
			pr.Error("%s\n", err.Error())
		}
	}

	for cnt, h := range resultList {

		title := fmt.Sprintf("%d:  %s (%s)", cnt+1, h.PrettyPrintRedir(cnt), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Fprintln(out, title)
		fmt.Fprintln(out, strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Fprintln(out)

		body, err := io.ReadAll(h.response.Body)
		if err != nil {
			pr.Errorln("%s", err)
		} else {
			fmt.Fprintln(out, at.Bold("Content:"))
			fmt.Fprintln(out, at.Bold(strings.Repeat(at.FrameOHLine, 8)))
			fmt.Fprintf(out, "\n%+v\n", string(body))
		}

		fmt.Fprintln(out)
	}
}
