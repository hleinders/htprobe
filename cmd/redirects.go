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

type RedirectFlags struct {
	showResponseHeader, showRequestHeader    bool
	showResponseCookies, showResponseCert    bool
	allHops, showContent, showRequestCookies bool
	displaySingleHeader, displaySingleCookie []string
}

var redirectFlags RedirectFlags

var redirectShortDesc = "Follows and shows the redirect chain of a http request"

// redirectsCmd represents the redirects command
var redirectsCmd = &cobra.Command{
	Use:     "redirects <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"rd", "redir", "redirect"},
	Short:   redirectShortDesc,
	Long: makeHeader(lowerAppName+" redirects: "+redirectShortDesc) + `With command 'redirects', the redirect chain of a http
request is shown. Every hop of this chain is displayed with the status code.
If the request is done via SSL and he certificate is invalid for some reason,
you may use the '-t|--trust' flag to force the connection to be trusted.
You can also display details like headers or cookies.

Flags marked with '***' may be used multiple times.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecRedirects(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(redirectsCmd)

	// flags
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showResponseCookies, "show-cookies", "c", false, "show response cookies")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showResponseCert, "show-cert", "C", false, "show certificate(s)")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showResponseHeader, "response-headers", "H", false, "show response headers")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showRequestHeader, "request-headers", "R", false, "show request headers")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showRequestCookies, "request-cookies", "Z", false, "show request cookies")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.showContent, "show-content", "O", false, "show content of last hop (prints to stderr)")
	redirectsCmd.Flags().BoolVarP(&redirectFlags.allHops, "all", "a", false, "show all details")

	// Parameter
	redirectsCmd.Flags().StringSliceVarP(&redirectFlags.displaySingleHeader, "display-header", "S", nil, "show only response header `FOOBAR`; ***")
	redirectsCmd.Flags().StringSliceVarP(&redirectFlags.displaySingleCookie, "display-cookie", "D", nil, "show only response cookie `SNAFU`; ***")

	// SingleHeader implies show ResponseHeaders:
	redirectFlags.showResponseHeader = redirectFlags.showResponseHeader || redirectsCmd.Flags().Changed("header")
}

func ExecRedirects(cmd *cobra.Command, args []string) {
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

		// handle the request
		hops, err = follow(&newReq, &connSet)
		if err != nil {
			pr.Error(err.Error())
		}

		// display results
		prettyPrintChain(hops)

		if redirectFlags.showContent {
			lastHop := hops[len(hops)-1]
			body, err := io.ReadAll(lastHop.response.Body)
			if err != nil {
				pr.Errorln("%s", err)
			} else {
				fmt.Fprintln(os.Stderr)
				fmt.Fprintln(os.Stderr, at.Bold("Content:"))
				fmt.Fprintln(os.Stderr, at.Bold(strings.Repeat(at.FrameOHLine, 8)))
				fmt.Fprintf(os.Stderr, "\n%+v\n", string(body))
			}
		}
	}
}

func prettyPrintChain(resultList []WebRequestResult) {
	var lastStatusCode int

	numItem := len(resultList) - 1

	// handle first item:
	fmt.Println()
	first := resultList[0]
	fmt.Println(first.PrettyPrintFirst())

	rdHandleHeaders(first, redirectFlags.allHops)
	rqHeaderDone = true
	rqCookiesDone = true

	// remember status
	lastStatusCode = first.response.StatusCode

	// no do the remaining
	if numItem >= 1 {
		for i, h := range resultList[1:] {
			fmt.Println(h.PrettyPrintNormal(lastStatusCode))

			showResponse := (i == numItem-1) || redirectFlags.allHops
			rdHandleHeaders(h, showResponse)

			lastStatusCode = h.response.StatusCode
		}
	}

	// last status:
	fmt.Println(resultList[numItem].PrettyPrintLast())
	fmt.Println()
}

func rdHandleHeaders(result WebRequestResult, showResponse bool) {
	// Request stuff:
	// Request headers: May only occour on first hop
	if redirectFlags.showRequestHeader && !rqHeaderDone {
		chainPrintHeaders(htab, vbar, at.BulletChar, "Request Header:", result.request.Header)
	}

	if redirectFlags.showRequestCookies && !rqCookiesDone {
		chainPrintCookies(htab, vbar, at.BulletChar, "Request Cookies:", result.request.Cookies())
	}

	// Response stuff
	// Response certificates: May occour in all hops or only at last hop
	if redirectFlags.showResponseCert && showResponse {
		chainPrintCertificates(htab, vbar, at.BulletChar, "Certificate(s):", result.response.TLS)
	}

	// Response Headers: May occour in all hops or only at last hop
	if redirectFlags.showResponseHeader && showResponse {
		if len(redirectFlags.displaySingleHeader) == 0 {
			chainPrintHeaders(htab, vbar, at.BulletChar, "Response Header:", result.response.Header)
		} else {
			chainPrintHeaders(htab, vbar, at.BulletChar, "Selected Headers:", makeHeadersFromName(redirectFlags.displaySingleHeader, result.response.Header))
		}
	}

	// Response cookies: May occour in all hops or only at last hop
	if redirectFlags.showResponseCookies && showResponse {
		if len(redirectFlags.displaySingleHeader) == 0 {
			chainPrintCookies(htab, vbar, at.BulletChar, "Stored Cookies:", result.cookieLst)
		} else {
			chainPrintCookies(htab, vbar, at.BulletChar, "Selected Cookies:", makeCookiesFromNames(result.cookieLst, cookieFlags.displaySingleCookie))
		}
	}
}
