package cmd

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

// Return values
const (
	OK = iota
	ErrUndef
	ErrNoArg
	ErrCookieJar
	ErrRequest
	ErrResponse
	ErrResolve
	ErrGetFlag
	ErrTimeFmt
	ErrNoURL
	ErrFileIO
	ErrNoFile
)

const (
	AppVersion               = "1.0 (2024-03-06)"
	Author                   = "Harald Leinders <harald@leinders.de>"
	DefaultConnectionTimeout = 3
	MaxRedirects             = 25
	MaxHeaderLen             = 30
)

type RequestMethod struct {
	name      string
	needsBody bool
}

type ConnectionSetup struct {
	timeOut       time.Duration
	proxy         string
	trust         bool
	follow        bool
	acceptCookies bool
	cookieJar     *cookiejar.Jar
}

type WebRequest struct {
	url       url.URL
	agent     string
	lang      string
	method    string
	authUser  string
	authPass  string
	reqBody   string
	xhdrs     []string
	cookieLst []*http.Cookie
}

func (r WebRequest) String() string {
	return fmt.Sprintf("%s (%s)", r.url.String(), r.method)
}

type WebRequestResult struct {
	request   http.Request
	response  http.Response
	cookieJar *http.CookieJar
}

func (r WebRequestResult) String() string {
	return fmt.Sprintf("%s (%s)", r.request.URL.String(), r.response.Status)
}

func (r WebRequestResult) GetRequest() string {
	reqStr := r.request.URL.String()

	if rootFlags.resolve {
		reqStr = fmt.Sprintf("%s (%s)", reqStr, doResolve(r.request.URL.Hostname()))
	}

	return reqStr
}

var AllowedHttpMethods = []RequestMethod{
	{"GET", false},
	{"HEAD", false},
	{"POST", true},
	{"PUT", true},
	{"PATCH", true},
	{"DELETE", false},
	{"CONNECT", false},
	{"OPTIONS", false},
	{"TRACE", false},
}

var (
	agentString       = "Golang Request Checker v" + AppVersion
	rqHeaderDone      = false
	globalRequestBody string
	globalCcookieLst  []*http.Cookie
	hcont             string
	corner            string
	vbar              string
	htab              string
	indentHeader      string
	rarrow            string
)
