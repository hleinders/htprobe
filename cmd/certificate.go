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

var certificateShortDesc = "Analyzes and displays server certificates"

// certificateCmd represents the certificate command
var certificateCmd = &cobra.Command{
	Use:     "certificate <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"ct", "crt", "cert"},
	Short:   certificateShortDesc,
	Long: makeHeader("htprobe certificate: "+certificateShortDesc) + `With 'htprobe certificate <URL>' the server certificate of URL
is shown. If the certifiace is invalid for some reason and the
connection is declined, you may force the connection with the
'-t|--trust' flag to force the connection to be trusted.
You may pass the '-f|--follow' flag to follow redirects. In this case,
the certificate can be displayed in any hop with the '-a|--all' flag.

Flags marked with '***' may be used multiple times.`,
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

func displayCertificates(indent, frameChar, mark, titleMsg string, tls *tls.ConnectionState) {
	var msgSAN string
	var nameFound bool
	var displayName string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if tls != nil {
		cert := tls.PeerCertificates
		c0 := cert[0]
		commonName := strings.ToLower(strings.TrimSpace(c0.Subject.CommonName))
		serverName := strings.ToLower(strings.TrimSpace(tls.ServerName))
		subjectANs := c0.DNSNames
		validUntil := c0.NotAfter

		displayName = commonName
		if commonName == serverName {
			displayName = at.Green(commonName)
			nameFound = true
		} else if found := findInSlice(subjectANs, serverName); found {
			subjectANs = markGreenInSlice(subjectANs, serverName)
			nameFound = true
		}

		if len(subjectANs) > 0 {
			msgSAN = strings.Join(subjectANs, ", ")
		} else {
			msgSAN = "None"
		}

		if !nameFound {
			displayName = at.Red(commonName)
			msgSAN = at.Red(msgSAN)
		}

		// print cert
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s CN:          %s", mark, displayName))

		// print sans
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  SANs:        %s", shorten(rootFlags.long, screenWidth-25, msgSAN)))

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

func chainPrintCertificates(indent, frameChar, mark, titleMsg string, tls *tls.ConnectionState) {
	var msgSAN, msgCaChain string
	var caChain []string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if tls != nil {
		cert := tls.PeerCertificates
		c0 := cert[0]
		commonName := strings.ToLower(strings.TrimSpace(c0.Subject.CommonName))
		subjectANs := c0.DNSNames
		validUntil := c0.NotAfter

		for _, c := range cert {
			caChain = append(caChain, c.Subject.CommonName)
		}

		msgCaChain = strings.Join(caChain, " <<< ")

		if len(subjectANs) > 0 {
			msgSAN = strings.Join(subjectANs, ", ")
		} else {
			msgSAN = "None"
		}

		// print cert
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s CN:          %s", mark, commonName))

		// print sans
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  SANs:        %s", shorten(rootFlags.long, screenWidth-20, msgSAN)))

		// validity
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Valid until: %s", colorValidity(validUntil)))

		// print chain
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  CA-Chain:    %s", shorten(rootFlags.long, screenWidth-20, msgCaChain)))
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

	fmt.Println()

	for cnt, h := range resultList {

		title := fmt.Sprintf("%d:  %s (%s)", cnt+1, h.PrettyPrintRedir(cnt), colorStatus(h.response.StatusCode))
		titleLen := len(stripColorCodes(title))

		fmt.Println(title)
		fmt.Println(strings.Repeat(at.FrameOHLine, titleLen))
		fmt.Println()
		displayCertificates(indentHeader, "", at.BulletChar, "Certificate(s):", h.response.TLS)
		fmt.Println()
	}
}
