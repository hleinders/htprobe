/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"crypto/tls"
	"fmt"
	"strings"

	at "github.com/hleinders/AnsiTerm"
	"github.com/spf13/cobra"
)

type CertificateFlags struct {
	follow bool
}

var certificateFlags CertificateFlags

// certificateCmd represents the certificate command
var certificateCmd = &cobra.Command{
	Use:     "certificate",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"crt", "cert"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecCertificate(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(certificateCmd)

	// flags
	certificateCmd.Flags().BoolVarP(&certificateFlags.follow, "follow", "f", false, "show all response cookies")
}

func ExecCertificate(cmd *cobra.Command, args []string) {
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
		if certificateFlags.follow {
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
		prettyPrintCertificates(hops)
	}
}

func chainPrintCertificates(indent, frameChar, mark, titleMsg string, tls *tls.ConnectionState) {
	var msgSAN string
	// var chain []string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if tls != nil {
		// chain := "CA-Chain: "
		cert := tls.PeerCertificates
		c0 := cert[0]
		commonName := c0.Subject.CommonName
		subjectANs := c0.DNSNames
		validUntil := c0.NotAfter

		if len(subjectANs) > 0 {
			msgSAN = strings.Join(subjectANs, ", ")
		} else {
			msgSAN = "None"
		}

		// print cert
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s CN:          %s", mark, commonName))

		// print sans
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  SANs:        %s", msgSAN))

		// validity
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Valid until: %s", colorValidity(validUntil)))

		// print chain
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  CA-Chain:    %s", commonName))
		for _, c := range cert[1:] {
			// chain = append(chain, c.Subject.CommonName)
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("               %s  %s", at.Larrow, c.Subject.CommonName))
		}

	} else {
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s", mark, "(None)"))
	}

	fmt.Printf("%s%s\n", indent, frameChar)
}

// func displayCert(contStr string, resp *http.Response) {
// 	var chain []string
// 	var msg, msgSAN string

// 	if resp.TLS != nil {
// 		// chain := "CA-Chain: "
// 		cert := resp.TLS.PeerCertificates
// 		c0 := cert[0]
// 		commonName := c0.Subject.CommonName
// 		subjectANs := c0.DNSNames
// 		validUntil := c0.NotAfter

// 		if len(subjectANs) > 0 {
// 			msgSAN = "[" + strings.Join(subjectANs, ", ") + "]"
// 		} else {
// 			msgSAN = "None"
// 		}

// 		for _, c := range cert[1:] {
// 			chain = append(chain, c.Subject.CommonName)
// 		}
// 		pprint(contStr, bold("Certificate(s):\n"))
// 		msg = fmt.Sprintf("CN: %s (SAN: %s)\n", commonName, msgSAN)
// 		pprint(contStr, msg)
// 		msg = fmt.Sprintf("CA-Chain: %s\n", strings.Join(chain, " "+larrow+" "))
// 		pprint(contStr, msg)
// 		msg = fmt.Sprintf("Valid until: %s\n", validUntil)
// 		pprint(contStr, msg)

// 		fmt.Printf("%s%s\n", frameHztab, contStr)
// 	}
// }

func prettyPrintCertificates(resultList []WebRequestResult) {
	numItem := len(resultList) - 1

	for _, h := range resultList {

		title := fmt.Sprintf("%s (%s)", h.PrettyPrintFirst(), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Println(title)
		fmt.Println(strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Println()
		chainPrintCertificates(indentHeader, "", at.BulletChar, "Certificate(s):", h.response.TLS)
		// chainPrintCertificatesChain(indentHeader, "", "", "Certificate(s):", h.response.TLS)
		fmt.Println()
	}

	// last status:
	resultList[numItem].PrettyPrintLast()
}
