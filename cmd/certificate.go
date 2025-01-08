/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	at "github.com/hleinders/AnsiTerm"

	"github.com/spf13/cobra"
)

type CertificateFlags struct {
	follow             bool
	showDetails        bool
	showValidatedChain bool
}

var certificateFlags CertificateFlags

var certificateShortDesc = "Analyzes and displays server certificates"

// certificateCmd represents the certificate command
var certificateCmd = &cobra.Command{
	Use:     "certificate <URL> [<URL> ...]",
	Args:    cobra.MinimumNArgs(1),
	Aliases: []string{"ct", "crt", "cert"},
	Short:   certificateShortDesc,
	Long: makeHeader(lowerAppName+" certificate: "+certificateShortDesc) + `With command 'certificate', the server certificate of URL
is shown. If the certifiace is invalid for some reason and the
connection is declined, you may force the connection with the
'-t|--trust' flag to force the connection to be trusted.
You may pass the '-f|--follow' flag to follow redirects.
Any server should send a full certificate chain (peer chain).
A client may prefer its own locally detected and verified chain,
wich can be displayed with the '-V|--validated-chain' flag.
This sometimes hides server misconfigurations.

Flags marked with '***' may be used multiple times.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecCertificate(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(certificateCmd)

	// flags
	certificateCmd.Flags().BoolVarP(&certificateFlags.follow, "follow", "f", false, "show response cookies for all hops")
	certificateCmd.Flags().BoolVarP(&certificateFlags.showDetails, "show-details", "s", false, "show certificate details")
	certificateCmd.Flags().BoolVarP(&certificateFlags.showValidatedChain, "validated-chain", "V", false, "display client side verified certificate chain")
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
		newReq.url, err = checkURL(rawURL, true)
		check(err, ErrNoURL)

		// handle the request(s)
		if certificateFlags.follow {
			hops, err = follow(&newReq, &connSet)
			if err != nil {
				pr.Error("%s", err.Error())
			}
		} else {
			hops, err = noFollow(&newReq, &connSet)
			if err != nil {
				pr.Error("%s", err.Error())
			}

		}

		// display results
		prettyPrintCertificates(hops)
	}
}

type Cert struct {
	commonName        string
	rawCName          string
	subjectANs        []string
	validFrom         time.Time
	validUntil        time.Time
	organization      string
	organizationUnits string
	country           string
	isCA              bool
	issuerName        string
	issuerOrg         string
	issuerOU          string
	issuerCountry     string
}

func makeCert(rawCert *x509.Certificate) Cert {
	var c Cert

	c.rawCName = rawCert.Subject.CommonName
	c.commonName = strings.ToLower(strings.TrimSpace(c.rawCName))
	c.subjectANs = rawCert.DNSNames
	c.validFrom = rawCert.NotBefore
	c.validUntil = rawCert.NotAfter
	c.isCA = rawCert.IsCA

	if list := rawCert.Subject.Organization; len(list) > 0 {
		c.organization = strings.Join(list, ", ")
	} else {
		c.organization = "(not available)"
	}

	if list := rawCert.Subject.OrganizationalUnit; len(list) > 0 {
		c.organizationUnits = strings.Join(list, ", ")
	}

	if list := rawCert.Subject.Country; len(list) > 0 {
		c.country = strings.Join(list, ", ")
	}

	c.issuerName = rawCert.Issuer.CommonName

	if list := rawCert.Issuer.Organization; len(list) > 0 {
		c.issuerOrg = strings.Join(list, ", ")
	} else {
		c.issuerOrg = "(not available)"
	}

	if list := rawCert.Issuer.OrganizationalUnit; len(list) > 0 {
		c.issuerOU = strings.Join(list, ", ")
	}

	if list := rawCert.Issuer.Country; len(list) > 0 {
		c.issuerCountry = strings.Join(list, ", ")
	}

	return c
}

func displayCertChain(count int, title, fmtString, indent, frameChar string, chain []*x509.Certificate) {
	commonName := strings.ToLower(chain[0].Subject.CommonName)
	fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  %-9s [%d] %s", title, count, commonName))

	for _, c := range chain[1:] {
		chainName := c.Subject.CommonName
		if len(chainName) == 0 {
			break
		}
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("                %s  %s (%s)", at.Larrow, chainName, strings.Join(c.Subject.Organization, ", ")))
	}

}

func displayCertificates(indent, frameChar, mark, titleMsg string, tls *tls.ConnectionState) {
	var msgSAN string
	var found, nameFound bool
	var displayName, heading string
	var c0 Cert

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if tls != nil {
		// prepare certs:
		peers := tls.PeerCertificates // default
		verifiedChains := tls.VerifiedChains

		pr.Debug("Verified Chain: %v\n", tls.VerifiedChains)
		pr.Debug("Peers: %v\n", peers)

		c0 = makeCert(peers[0])

		serverName := strings.ToLower(strings.TrimSpace(tls.ServerName))
		displayName = c0.commonName

		if matchGlob(c0.commonName, serverName) {
			displayName = at.Green(c0.commonName)
			nameFound = true
		} else if found = findGlobInSlice(c0.subjectANs, serverName); found {
			c0.subjectANs = markGreenInSlice(c0.subjectANs, serverName)
			nameFound = true
		}

		if len(c0.subjectANs) > 0 {
			msgSAN = strings.Join(c0.subjectANs, ", ")
		} else {
			msgSAN = "None"
		}

		if !nameFound {
			displayName = at.Red(c0.commonName)
			msgSAN = at.Red(msgSAN)
		}

		// print cert
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s CN:           %s", mark, shorten(rootFlags.long, screenWidth-25, displayName)))
		resetColor()

		if certificateFlags.showDetails {
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Organization: %s", c0.organization))
			if c0.organizationUnits != "" {
				fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Unit:         %s", c0.organizationUnits))
			}
			if c0.country != "" {
				fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Country:      %s", c0.country))
			}
		}

		// print sans
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  SANs:         %s", shorten(rootFlags.long, screenWidth-25, msgSAN)))
		resetColor()

		// print issuer
		if certificateFlags.showDetails {
			fmt.Println()
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Issuer:       %s", shorten(rootFlags.long, screenWidth-25, c0.issuerName)))
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Organization: %s", c0.issuerOrg))
			if c0.issuerOU != "" {
				fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Unit:         %s", c0.issuerOU))
			}
			if c0.issuerCountry != "" {
				fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Country:      %s", c0.issuerCountry))
			}
		}

		// validity
		if certificateFlags.showDetails {
			fmt.Println()
			fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Valid from:   %s", c0.validFrom))
		}
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Valid until:  %s", colorValidity(c0.validUntil)))

		// print peer chain
		if certificateFlags.showDetails {
			fmt.Println()
			peerType := at.Green("sent by peer")
			if len(peers) < 2 {
				peerType = at.Red("incomplete")
			}
			if connSet.trust {
				peerType = at.Yellow("trust forced")
			}
			if c0.isCA {
				peerType = at.Red("selfsigned")
			}

			fmt.Printf(fmtString, indent, frameChar, "  Certificate Chain ("+peerType+"):")
			heading = ""
		} else {
			heading = "CA-Chain:"

		}
		displayCertChain(0, heading, fmtString, indent, frameChar, peers)

		// print verified chain
		if certificateFlags.showDetails && certificateFlags.showValidatedChain {
			fmt.Println()
			if len(verifiedChains) > 0 {
				fmt.Printf(fmtString, indent, frameChar, "  Verified Chain(s) ("+at.Green("checked by client")+"):")
				for k, crt := range verifiedChains {
					displayCertChain(k, heading, fmtString, indent, frameChar, crt)
				}
			} else {
				fmt.Printf(fmtString, indent, frameChar, "  Verified Chain(s): None")
			}
		}
	} else {
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s", mark, "(None)"))
	}

	fmt.Printf("%s%s\n", indent, frameChar)
}

// function used by redirects module
func chainPrintCertificates(indent, frameChar, mark, titleMsg string, tls *tls.ConnectionState) {
	var msgSAN, msgCaChain string
	var caChain []string

	fmtString := "%s%s   %s\n"
	fmt.Printf(fmtString, indent, frameChar, at.Bold(titleMsg))

	if tls != nil {
		certs := tls.PeerCertificates
		c0 := certs[0]
		commonName := strings.ToLower(strings.TrimSpace(c0.Subject.CommonName))
		subjectANs := c0.DNSNames
		validUntil := c0.NotAfter

		for _, c := range certs {
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
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  SANs:        %s", shorten(rootFlags.long, screenWidth-28, msgSAN)))

		// validity
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  Valid until: %s", colorValidity(validUntil)))

		// print chain
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("  CA-Chain:    %s", shorten(rootFlags.long, screenWidth-28, msgCaChain)))
	} else {
		fmt.Printf(fmtString, indent, frameChar, fmt.Sprintf("%s %s", mark, "(None)"))
	}

	fmt.Printf("%s%s\n", indent, frameChar)
}

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
