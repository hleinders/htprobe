package cmd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	at "github.com/hleinders/AnsiTerm"
)

func check(e error, rcode int) {
	if e != nil {
		fmt.Fprintf(os.Stderr, at.Red("*** Error: %+v\n"), e)
		os.Exit(rcode)
	}
}

func makeHeader(str string) string {
	// return "\n" + at.Bold(str) + "\n\n"
	return "\n" + at.Yellow(str) + "\n\n"
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

func markGreenInSlice(slice []string, str string) []string {
	var result []string
	var tmp string

	s := strings.TrimSpace(str)
	for _, item := range slice {
		if tmp = strings.TrimSpace(item); tmp == s {
			tmp = at.Green(tmp)
		}
		result = append(result, tmp)
	}

	return result
}

func shorten(disable bool, length int, str string) string {
	if len(str) > length && !disable {
		return str[:length-3] + "..."
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

func colorValidity(validUntil time.Time) string {
	now := time.Now()
	diff := validUntil.Sub(now).Hours() / 24
	str := validUntil.String()

	if diff < 0 {
		return at.Red(str)
	}

	if diff < 30 {
		return at.Yellow(str)
	}

	return at.Green(str)
}

func stripColorCodes(str string) string {
	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

	var re = regexp.MustCompile(ansi)

	return re.ReplaceAllString(str, "")
}

// This function parses an string of the form "name: value" to a cookie
func getCookieFromString(raw string) (http.Cookie, error) {
	var c http.Cookie
	var err error

	r := strings.SplitN(raw, ":", 2)
	if len(r) == 2 {
		c = http.Cookie{Name: strings.TrimSpace(r[0]), Value: strings.TrimSpace(r[1])}
	} else {
		err = fmt.Errorf("could not parse: %s", raw)
	}

	return c, err
}

func deleteCookieFromList(cookie *http.Cookie, list []*http.Cookie) []*http.Cookie {
	var result []*http.Cookie

	for _, c := range list {
		if c.Name != cookie.Name {
			result = append(result, c)
		}
	}

	return result
}

func findCookieInList(cookie *http.Cookie, list []*http.Cookie) bool {
	for _, c := range list {
		if c.Name == cookie.Name {
			return true
		}
	}

	return false
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
		if client.Jar != nil {
			result.cookieLst = client.Jar.Cookies(resp.Request.URL)
		}
	} else {
		pr.Debug("Error is: %s\n", reflect.TypeOf(errReq))
		pr.Debug("Error details: %+v\n", errors.Unwrap(errReq))
		te := errors.Unwrap(errReq)
		pr.Debug("Unwrapped error is: %s\n", reflect.TypeOf(te))
		pr.Debug("Unwrapped error details: %+v\n", errors.Unwrap(te))
	}

	// check if cookie went to jar:
	for _, c := range result.cookieLst {
		if findCookieInList(c, wr.cookieLst) {
			wr.cookieLst = deleteCookieFromList(c, wr.cookieLst)
		}
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

func noFollow(wr *WebRequest, cs *ConnectionSetup) ([]WebRequestResult, error) {
	var resultList []WebRequestResult

	// init client
	hc := initClient(cs)

	// initial and only request
	result, err := doRequest(hc, wr)
	check(err, ErrRequest)

	// add to list:
	resultList = append(resultList, result)

	return resultList, err
}

func checkURL(rawURL string) (url.URL, error) {

	// has arg a protocol?
	rx := regexp.MustCompile(`(?i)^https?://`)
	if !rx.Match([]byte(rawURL)) {
		pr.Debug("Added protocol prefix to %s.\n", rawURL)
		rawURL = "http://" + rawURL
	}

	// is arg an url?
	pr.Debug("Raw URL: %s\n", rawURL)

	u, e := url.ParseRequestURI(rawURL)
	return *u, e
}
