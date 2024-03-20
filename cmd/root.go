/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"bufio"
	"fmt"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	at "github.com/hleinders/AnsiTerm"
	cp "github.com/hleinders/colorprint"
	"golang.org/x/net/publicsuffix"

	"github.com/spf13/cobra"
)

type RootFlags struct {
	debug, verbose                        bool
	noColor, noFancy, ascii               bool
	resolve, long                         bool
	agent, reqLang, httpMethod            string
	authUser, authPass                    string
	cookieFile, bodyFile, headerFile      string
	cookieValues, bodyValues, xtraHeaders []string
}

var (
	rootFlags   RootFlags
	pr          *cp.Printer
	connSet     ConnectionSetup
	connTimeout int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "htprobe",
	Version: AppVersion,
	Short:   "A http request analyzing and debugging tool",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PersistentPreRun: PersistentPreRun,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// flags
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.verbose, "verbose", "v", false, "set verbose mode")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.debug, "debug", false, "set debug mode")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.ascii, "ascii", false, "use ascii chars only")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.noColor, "no-color", false, "do not use colors")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.noFancy, "no-fancy", false, "combines no color and ascii mode")
	rootCmd.PersistentFlags().BoolVarP(&connSet.trust, "trust", "t", false, "trust selfsigned certificates")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.resolve, "resolve", false, "resolve host names")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.long, "long", "l", false, "long output, don't shorten results (header, cookies etc.)")
	rootCmd.PersistentFlags().BoolVarP(&connSet.acceptCookies, "accept-cookies", "a", false, "accept response cookies")

	// Parameter
	rootCmd.PersistentFlags().StringVarP(&rootFlags.authUser, "user", "u", "", "`user` (basic auth)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.authPass, "pass", "p", "", "`password` (basic auth)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.agent, "agent", agentString, "user agent")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.reqLang, "lang", "L", "", "set `language` header for request")
	rootCmd.PersistentFlags().StringVarP(&connSet.proxy, "proxy", "P", "", "set `host` as proxy")
	rootCmd.PersistentFlags().IntVarP(&connTimeout, "timeout", "T", DefaultConnectionTimeout, "connection `time`out in seconds (0=disable, <=3600)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.httpMethod, "method", "m", "GET", "http request `method` (see RFC 7231 section 4.3.)")
	rootCmd.PersistentFlags().StringSliceVarP(&rootFlags.cookieValues, "cookie", "c", nil, "set request cookie (fmt: `name:value`)\n(Can be used multiple times)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.cookieFile, "cookie-file", "C", "", "read cookies from `file` (fmt: lines of 'name:value')")
	rootCmd.PersistentFlags().StringSliceVarP(&rootFlags.bodyValues, "body", "b", nil, "add `entry` to request body where needed (e.g. POST)\n(Can be used multiple times)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.bodyFile, "body-file", "B", "", "read request body from `file`")
	rootCmd.PersistentFlags().StringSliceVarP(&rootFlags.xtraHeaders, "header", "r", nil, "pass `header` to request (fmt: 'Name: Value')\n(Can be used multiple times)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.headerFile, "header-file", "R", "", "read extra request headers from `file` (fmt: lines of 'name:value')")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.MarkFlagsRequiredTogether("user", "pass")
}

func PersistentPreRun(cmd *cobra.Command, args []string) {
	var err error
	var cookieStringList, bodyList, headerStringList []string

	// handle fancy stuff
	color.NoColor = rootFlags.noColor || at.NoColor()

	// handle charset
	if rootFlags.ascii {
		at.AsciiChars()
	}

	// set up fancy chars:
	hcont = at.FrameTLineL + at.FrameHLine + at.FrameHLine
	corner = at.FrameCloseL + at.FrameHLine + at.FrameHLine
	vbar = at.FrameVLine
	htab = strings.Repeat(" ", 7)
	indentHeader = strings.Repeat(" ", 2)
	rarrow = at.Harrow

	// init printer
	pr = cp.NewPrinter()
	pr.SetVerbose(rootFlags.verbose)
	pr.SetDebug(rootFlags.debug)

	connSet.timeOut, err = time.ParseDuration(fmt.Sprintf("%ds", connTimeout))
	check(err, ErrTimeFmt)
	if connSet.acceptCookies {
		connSet.cookieJar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		check(err, ErrCookieJar)
	}

	//
	// Handle request headers:
	if rootFlags.xtraHeaders != nil {
		headerStringList = append(cookieStringList, rootFlags.xtraHeaders...)
	}

	if rootFlags.headerFile != "" {
		if cf, err0 := os.Open(rootFlags.headerFile); err0 == nil {
			defer cf.Close()

			r := bufio.NewScanner(cf)
			for r.Scan() {
				headerStringList = append(headerStringList, r.Text())
			}

			if err1 := r.Err(); err1 != nil {
				check(err1, ErrFileIO)
			}
		} else {
			check(err0, ErrNoFile)
		}
	}
	globalHeaderList = headerStringList
	pr.Debug("Request headers from flags: \n%s\n", globalCookieLst)

	//
	// Handle request cookies:
	if rootFlags.cookieValues != nil {
		cookieStringList = append(cookieStringList, rootFlags.cookieValues...)
	}

	if rootFlags.cookieFile != "" {
		if cf, err0 := os.Open(rootFlags.cookieFile); err0 == nil {
			defer cf.Close()

			r := bufio.NewScanner(cf)
			for r.Scan() {
				cookieStringList = append(cookieStringList, r.Text())
			}

			if err1 := r.Err(); err1 != nil {
				check(err1, ErrFileIO)
			}
		} else {
			check(err0, ErrNoFile)
		}
	}

	if len(cookieStringList) > 0 {
		for _, ci := range cookieStringList {
			c, err := getCookieFromString(ci)
			if err != nil {
				continue
			}

			globalCookieLst = append(globalCookieLst, &c)
		}
	}
	pr.Debug("Request cookies from flags: \n%s\n", globalCookieLst)

	//
	// Handle request body:
	if rootFlags.bodyValues != nil {
		bodyList = append(bodyList, rootFlags.bodyValues...)
	}

	if rootFlags.bodyFile != "" {
		if bf, err0 := os.Open(rootFlags.bodyFile); err0 == nil {
			defer bf.Close()

			r := bufio.NewScanner(bf)
			for r.Scan() {
				bodyList = append(bodyList, r.Text())
			}

			if err1 := r.Err(); err1 != nil {
				check(err1, ErrFileIO)
			}
		} else {
			check(err0, ErrNoFile)
		}
	}

	globalRequestBody = strings.Join(bodyList, "\n")
	pr.Debug("Request body from flags: \n%s\n", globalRequestBody)
}
