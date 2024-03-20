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
	showCookiesDetails                    bool
	displaySingleHeader                   []string
}

var redirectFlags RedirectFlags

// redirectsCmd represents the redirects command
var redirectsCmd = &cobra.Command{
	Use:     "redirects <URL>",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"rd", "redir"},
	Short:   "Follows and shows the redirect chain of a http request",
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
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showContent, "show-content", "O", false, "show content of last hop\n(prints to stderr)")
	redirectsCmd.Flags().BoolVar(&redirectFlags.allHops, "all", false, "show all details in every hop")

	// Parameter
	redirectsCmd.Flags().StringSliceVarP(&redirectFlags.displaySingleHeader, "header", "H", nil, "show only response header `FOOBAR`\n(Can be used multiple times)")

	// details imply showing
	redirectFlags.showCookies = redirectFlags.showCookies || redirectFlags.showCookiesDetails

	// SingleHeader implies show ResponseHeaders:
	redirectFlags.showResponseHeader = redirectFlags.showResponseHeader || redirectsCmd.Flags().Changed("header")
}

func ExecRedirects(cmd *cobra.Command, args []string) {
	var hops []WebRequestResult

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
		hops, err = follow(&newReq, &connSet)
		if err != nil {
			pr.Error(err.Error())
		}

		// display results
		prettyPrintChain(hops)
		fmt.Println()
	}
}

func prettyPrintChain(resultList []WebRequestResult) {
	var lastStatusCode int
	// var lastStatus string

	numItem := len(resultList) - 1

	// handle first item:
	first := resultList[0]
	fmt.Println(first.PrettyPrintFirst())

	rdHandleHeaders(first, redirectFlags.allHops)
	rqHeaderDone = true

	// remember status
	lastStatusCode = first.response.StatusCode

	// no do the remaining
	if numItem > 1 {
		for i, h := range resultList[1:] {
			fmt.Println(h.PrettyPrintNormal(lastStatusCode))

			showResponse := (i == numItem-1) || redirectFlags.allHops
			rdHandleHeaders(h, showResponse)

			lastStatusCode = h.response.StatusCode
		}
	}

	// last status:
	fmt.Println(resultList[numItem].PrettyPrintLast())
}

func rdHandleHeaders(result WebRequestResult, showRespons bool) {
	// Request headers: May only occour on first hop
	if redirectFlags.showRequestHeader && !rqHeaderDone {
		chainPrintHeaders(htab, vbar, at.BulletChar, "Request Header:", result.request.Header)
	}

	// Response Headers: May occour in all hops or only at last hop
	if redirectFlags.showResponseHeader && showRespons {
		if len(redirectFlags.displaySingleHeader) == 0 {
			chainPrintHeaders(htab, vbar, at.BulletChar, "Response Header:", result.response.Header)
		} else {
			chainPrintHeaders(htab, vbar, at.BulletChar, "Selected Headers:", makeHeadersFromName(redirectFlags.displaySingleHeader, result.response.Header))
		}
	}
}
