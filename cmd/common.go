package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	at "github.com/hleinders/AnsiTerm"
)

func check(e error, rcode int) {
	if e != nil {
		fmt.Fprintf(os.Stderr, at.Red("*** Error: %+v\n"), e)
		os.Exit(rcode)
	}
}

func splitFirst(str, sep string) (string, string) {
	s := strings.SplitN(str, sep, 2)

	if len(s) < 2 {
		return s[0], ""
	}

	return s[0], strings.Join(s[1:], "")
}

func findInSlice(slice []string, str string) bool {
	s := strings.TrimSpace(str)
	for _, item := range slice {
		if strings.TrimSpace(item) == s {
			return true
		}
	}

	return false
}

func shorten(disable bool, str string) string {
	if len(str) > MaxHeaderLen && !disable {
		return str[:MaxHeaderLen-3] + "..."
	}

	return str
}

func doResolve(host string) string {
	sip, e := net.LookupHost(host)
	check(e, ErrResolve)
	return strings.Join(sip, ", ")
}

func getMethodNames() []string {
	var names []string

	for _, n := range AllowedHttpMethods {
		names = append(names, n.name)
	}

	return names
}

func methodNeedsBody(method string) bool {
	for _, n := range AllowedHttpMethods {
		if n.name == method {
			return n.needsBody
		}
	}

	return false
}

func colorStatus(stat int) string {

	switch statStr := strconv.Itoa(stat); {
	case stat < 100:
		return statStr
	case stat < 200:
		return at.Cyan(statStr)
	case stat < 300:
		return at.Green(statStr)
	case stat < 400:
		return at.Yellow(statStr)
	case stat < 500:
		return at.Red(statStr)
	case stat < 600:
		return at.Magenta(statStr)
	default:
		return statStr
	}
}

// =================================== HTTP Request Functions ==================================
func initClient(cs *ConnectionSetup) *http.Client {
	var rdf func(req *http.Request, via []*http.Request) error

	tr := &http.Transport{}

	if cs.trust {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if cs.proxy != "" {
		tr.Proxy = func(*http.Request) (*url.URL, error) {
			return url.Parse(fmt.Sprintf("http://%s", cs.proxy))
		}
	}

	if !cs.follow {
		rdf = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		rdf = nil
	}

	hc := &http.Client{
		CheckRedirect: rdf,
		Transport:     tr,
		Timeout:       cs.timeOut,
	}

	if cs.acceptCookies {
		hc.Jar = cs.cookieJar
	}

	return hc
}

func doRequest(client *http.Client, wr *WebRequest) (WebRequestResult, error) {
	var rb io.Reader
	var result WebRequestResult

	if methodNeedsBody(wr.method) {
		rb = strings.NewReader(wr.reqBody)
	}

	req, err := http.NewRequest(wr.method, wr.url.String(), rb)
	if err != nil {
		return result, err
	}
	req.Header.Set("User-Agent", wr.agent)

	// reqLang?
	if wr.lang != "" {
		req.Header.Set("Accept-Language", wr.lang)
	}

	// Additional header
	if len(wr.xhdrs) > 0 {
		for _, s := range wr.xhdrs {
			n, v := splitFirst(s, ":")
			req.Header.Set(strings.TrimSpace(n), strings.TrimSpace(v))
		}
	}

	// Auth?
	if wr.authUser != "" && wr.authPass != "" {
		req.SetBasicAuth(wr.authUser, wr.authPass)
	}

	if len(wr.cookieLst) > 0 {
		for _, c := range wr.cookieLst {
			req.AddCookie(c)
		}
	}

	pr.Debug("Request:\n%+v\n", req)
	pr.Debug("Cookies:\n%+v\n", client.Jar)

	// handle request
	resp, errReq := client.Do(req)

	if errReq == nil {
		result.request = *req
		result.response = *resp
		result.cookieJar = &client.Jar
	}

	return result, errReq
}

func follow(wr *WebRequest, cs *ConnectionSetup) ([]WebRequestResult, error) {
	var resultList []WebRequestResult

	// init client
	hc := initClient(cs)

	// initial request
	result, err := doRequest(hc, wr)
	check(err, ErrRequest)

	// add to list:
	resultList = append(resultList, result)

	cnt := 0
	// repeat until no further redirect happens:
	for result.response.StatusCode >= 301 && result.response.StatusCode <= 399 {
		// detect next hop:
		rdURL, e := result.response.Location()
		check(e, ErrResponse)

		// update the request
		wr.url = *rdURL
		wr.method = result.response.Request.Method

		cnt++
		if cnt >= MaxRedirects {
			result.response.StatusCode = 999
			break
		}

		// next hop:
		result, err = doRequest(hc, wr)
		check(err, ErrRequest)

		// add to list:
		resultList = append(resultList, result)
	}

	return resultList, err
}
