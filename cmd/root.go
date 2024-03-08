/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"
	"net/http/cookiejar"
	"os"
	"time"

	cp "github.com/hleinders/colorprint"
	"github.com/spf13/cobra"
	"golang.org/x/net/publicsuffix"
)

type RootFlags struct {
	debug, verbose          bool
	noColor, noFancy, ascii bool
	resolve, full           bool
	agent, reqLang          string
	authUser, authPass      string
	cookie, cookieFile      string
	httpMethod              string
}

var (
	rootFlags   RootFlags
	pr          *cp.Printer
	connSet     ConnectionSetup
	connTimeout int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "rqcheck",
	Version: AppVersion,
	Short:   "A brief description of your application",
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
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.verbose, "verbose", "v", false, "set verbose mode")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.debug, "debug", false, "set debug mode")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.ascii, "ascii", false, "use ascii chars only")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.noColor, "no-color", false, "do not use colors")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.noFancy, "no-fancy", false, "combines no color and ascii mode")
	rootCmd.PersistentFlags().BoolVarP(&connSet.trust, "trust", "t", false, "trust selfsigned certificates")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.resolve, "resolve", false, "resolve host names")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.full, "full", "f", false, "show results uncut (header, cookies etc.)")
	rootCmd.PersistentFlags().BoolVarP(&connSet.acceptCookies, "accept-cookies", "a", false, "accept response cookies")

	// Parameter
	rootCmd.PersistentFlags().StringVarP(&rootFlags.authUser, "user", "u", "", "`user` (basic auth)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.authPass, "pass", "p", "", "`password` (basic auth)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.agent, "agent", agentString, "user agent")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.reqLang, "lang", "L", "", "set `language` header for request")
	rootCmd.PersistentFlags().StringVarP(&connSet.proxy, "proxy", "P", "", "set `host` as proxy")
	rootCmd.PersistentFlags().IntVarP(&connTimeout, "timeout", "T", DefaultConnectionTimeout, "conn. `time`out in seconds (0=disable, <=3600)")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.httpMethod, "method", "m", "GET", "http request `method` (see RFC 7231 section 4.3.)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.cookie, "cookie", "", "set request cookie (fmt: `name:value`)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.cookieFile, "cookie-file", "", "read cookies from `file` (fmt: lines of 'name:value')")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().MarkHidden("debug")
	rootCmd.MarkFlagsRequiredTogether("user", "pass")
}

func PersistentPreRun(cmd *cobra.Command, args []string) {
	var err error

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
}
