/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"strings"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type CookieFlags struct {
	follow              bool
	displaySingleCookie []string
}

var cookieFlags CookieFlags

// cookiesCmd represents the cookies command
var cookiesCmd = &cobra.Command{
	Use:     "cookies",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"ck", "cookie"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecCookies(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(cookiesCmd)

	// flags
	cookiesCmd.Flags().BoolVarP(&cookieFlags.follow, "follow", "f", false, "show all response cookies")

	// Parameter
	cookiesCmd.Flags().StringSliceVarP(&cookieFlags.displaySingleCookie, "show-cookie", "S", nil, "show only cookie `FOOBAR`\n(Can be used multiple times)")

}

func ExecCookies(cmd *cobra.Command, args []string) {
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
		if cookieFlags.follow {
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
		prettyPrintCookies(hops)
	}
}

func makeCookiesFromResponse(headers http.Header) []*http.Cookie {
	var cl []*http.Cookie

	if cookieHeader := headers.Values("Set-Cookie"); len(cookieHeader) > 0 {
		c := http.Cookie{Name: "Set-Cookie", Value: "xxxx"}
		cl = append(cl, &c)
	}
	return cl
}

func chainPrintCookies(indent, frameChar, mark, titleMsg string, cookieList []*http.Cookie) {
	// var headerVal string
	// var cookieKeys []string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if len(cookieList) > 0 {
		for _, c := range cookieList {
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s: %s", mark, c.Name, shorten(rootFlags.long, fullCookieValues(c))))
		}
	} else {
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s", mark, "(None)"))
	}
	fmt.Printf("%s%s\n", indent, frameChar)
}

func fullCookieValues(c *http.Cookie) string {
	fv := c.Value

	if c.Path != "" {
		fv = fmt.Sprintf("%s; path=%s", fv, c.Path)
	}

	if c.Domain != "" {
		fv = fmt.Sprintf("%s; domain=%s", fv, c.Domain)
	}

	return fv
}

// func fullCookieString(c *http.Cookie) string {
// 	return fmt.Sprintf("%s=%s", c.Name, fullCookieValues(c))
// }

func prettyPrintCookies(resultList []WebRequestResult) {
	numItem := len(resultList) - 1

	for _, h := range resultList {

		// result title
		title := fmt.Sprintf("%s (%s)", h.PrettyPrintFirst(), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Println(title)
		fmt.Println(strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Println()

		// print cookies
		ckHandleCookies(h)
		fmt.Println()
	}

	// last status:
	resultList[numItem].PrettyPrintLast()
}

func ckHandleCookies(result WebRequestResult) {
	if len(cookieFlags.displaySingleCookie) == 0 {
		// Request cookies from globalCookieList
		chainPrintCookies(indentHeader, "", at.BulletChar, "Request Cookies:", result.request.Cookies())

		// Set-Cookie in response headers?
		responseCookies := makeCookiesFromResponse(result.response.Header)

		if len(responseCookies) > 0 {
			chainPrintHeaders(indentHeader, "", at.BulletChar, at.Yellow("Cookie Store Request Detected:"), makeHeadersFromName([]string{"Set-Cookie"}, result.response.Header))
		}

		if result.cookieLst != nil {
			chainPrintCookies(indentHeader, "", at.BulletChar, "Stored Cookies:", result.cookieLst)
		}
		// } else {
		// 	chainPrintHeaders(indentHeader, "", at.BulletChar, "Selected Cookies:", makeCookiesFromName(headerFlags.displaySingleHeader, result.response.Header))
	}
}
