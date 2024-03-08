/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"
	"net/url"
	"regexp"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type RedirectFlags struct {
	showResponseHeader, showRequestHeader bool
	allHops                               bool
	showContent, showCert                 bool
	showCookies                           bool
	unsorted, showCookiesDetails          bool
	displaySingleHeader                   []string
}

var redirectFlags RedirectFlags

// redirectsCmd represents the redirects command
var redirectsCmd = &cobra.Command{
	Use:   "redirects <URL>",
	Args:  cobra.MinimumNArgs(1),
	Short: "Follows and shows the redirect chain of a http request",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecRedirects(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(redirectsCmd)

	// flags
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showCookies, "show-cookies", "c", false, "show all response cookies")
	redirectsCmd.Flags().BoolVar(&redirectFlags.showCookiesDetails, "cookie-details", false, "show cookie details")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showCert, "show-cert", "C", false, "show certificate(s)")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showResponseHeader, "response-header", "R", false, "show all response header")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showRequestHeader, "request-header", "Q", false, "show all request header")
	redirectsCmd.Flags().BoolVar(&redirectFlags.unsorted, "no-sort", false, "do not sort header")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showContent, "show-content", "O", false, "show content of last hop\n(prints to stderr)")
	redirectsCmd.Flags().BoolVar(&redirectFlags.allHops, "all", false, "show all details in every hop")

	// Parameter
	redirectsCmd.Flags().StringSliceVarP(&redirectFlags.displaySingleHeader, "header", "H", nil, "show response `header`\n(Can be used multiple times)")

	// details imply showing
	redirectFlags.showCookies = redirectFlags.showCookies || redirectFlags.showCookiesDetails
}

func ExecRedirects(cmd *cobra.Command, args []string) {
	// create template request:
	req := WebRequest{
		agent:     rootFlags.agent,
		lang:      rootFlags.reqLang,
		method:    rootFlags.httpMethod,
		authUser:  rootFlags.authUser,
		authPass:  rootFlags.authPass,
		reqBody:   globalRequestBody,
		xhdrs:     rootFlags.xtraHeaders,
		cookieLst: globalCcookieLst,
	}

	for _, rawURL := range args {
		// has arg a protocol?
		rx := regexp.MustCompile(`(?i)^https?://`)
		if !rx.Match([]byte(rawURL)) {
			pr.Debug("Added protocol prefix to %s.\n", rawURL)
			rawURL = "http://" + rawURL
		}

		// is arg an url?
		pr.Debug("Raw URL: %s\n", rawURL)
		newURL, err := url.ParseRequestURI(rawURL)
		check(err, ErrNoURL)

		newReq := req
		newReq.url = *newURL

		// handle the request
		hops, err := follow(&newReq, &connSet)
		if err != nil {
			pr.Error(err.Error())
		}

		// display results
		printChain(hops)
		fmt.Println()
	}
}

func printChain(resultList []WebRequestResult) {
	var lastStatusCode int
	var lastStatus string

	numItems := len(resultList)

	// handle first item:
	first := resultList[0]
	fmt.Printf(at.Bold("URL: %s  [%s]\n"), first.GetRequest(), first.request.Method)

	// remember status
	lastStatusCode = first.response.StatusCode
	lastStatus = first.response.Status

	// no do the remaining
	if numItems > 1 {
		for _, h := range resultList[1:] {
			fmt.Printf("%s%s (%s) %s  [%s] %s\n", htab, hcont, colorStatus(lastStatusCode), rarrow, h.request.Method, h.GetRequest())

			lastStatusCode = h.response.StatusCode
			lastStatus = h.response.Status
		}
	}

	// last status:
	fmt.Printf("%s%s (%s) %s  %s\n", htab, corner, colorStatus(lastStatusCode), rarrow, lastStatus)

}
