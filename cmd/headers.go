/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type HeaderFlags struct {
	follow              bool
	displaySingleHeader []string
}

var headerFlags HeaderFlags

var headerShortDesc = "Shows the request and response headers of a http request"

// headersCmd represents the headers command
var headersCmd = &cobra.Command{
	Use:     "headers <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"hd", "hdr", "head"},
	Short:   headerShortDesc,
	Long: makeHeader("htprobe headers: "+headerShortDesc) + `With 'htprobe headers <URL>', all request and response headers
are shown. You may pass the '-f|--follow' flag to follow redirects.
In this case, the headers can be displayed in any hop with the '-a|--all' flag.

Flags marked with '***' may be used multiple times.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecHeaders(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(headersCmd)

	// flags
	headersCmd.Flags().BoolVarP(&headerFlags.follow, "follow", "f", false, "show headers for all hops")

	// Parameter
	headersCmd.Flags().StringSliceVarP(&headerFlags.displaySingleHeader, "show-header", "S", nil, "show only response header `FOOBAR`; ***")
}

func ExecHeaders(cmd *cobra.Command, args []string) {
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
		prettyPrintHeaders(hops)
	}
}

func makeHeadersFromName(names []string, headers http.Header) http.Header {
	result := http.Header{}

	var tmp string

	for _, n := range names {
		if h := headers.Values(n); len(h) != 0 {
			tmp = strings.Join(h, ", ")
		} else {
			tmp = at.Yellow("N/A")
		}
		result.Add(n, tmp)
	}
	return result
}

func chainPrintHeaders(indent, frameChar, mark, titleMsg string, headerList http.Header) {
	var headerKeys []string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	for h := range headerList {
		headerKeys = append(headerKeys, h)
	}

	sort.Strings(headerKeys)

	for _, h := range headerKeys {
		headerStr := fmt.Sprintf("%s: %s", h, strings.Join(headerList.Values(h), ", "))
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s", mark, shorten(rootFlags.long, screenWidth-20, headerStr)))
	}
	fmt.Printf("%s%s\n", indent, frameChar)
}

func prettyPrintHeaders(resultList []WebRequestResult) {

	fmt.Println()

	for cnt, h := range resultList {

		title := fmt.Sprintf("%d:  %s (%s)", cnt+1, h.PrettyPrintRedir(cnt), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Println(title)
		fmt.Println(strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Println()
		hdHandleHeaders(h)
		fmt.Println()
	}
}

func hdHandleHeaders(result WebRequestResult) {
	if len(headerFlags.displaySingleHeader) == 0 {
		chainPrintHeaders(indentHeader, "", at.BulletChar, "Request Header:", result.request.Header)
		chainPrintHeaders(indentHeader, "", at.BulletChar, "Response Header:", result.response.Header)
	} else {
		chainPrintHeaders(indentHeader, "", at.BulletChar, "Selected Headers:", makeHeadersFromName(headerFlags.displaySingleHeader, result.response.Header))
	}
}
