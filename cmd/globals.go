/*
Copyright © 2024 Dr. Harald Leinders <harald@leinders.de>
*/
package cmd

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	at "github.com/hleinders/AnsiTerm"
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
	ErrNoMethod
)

const (
	AppName                  = "HtProbe"
	AppVersion               = "1.14 (2025-09-05)"
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
	cookieLst []*http.Cookie
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

func (r WebRequestResult) PrettyPrintFirst() string {
	return fmt.Sprintf(at.Bold("URL: %s  [%s]"), r.GetRequest(), r.request.Method)
}

func (r WebRequestResult) PrettyPrintRedir(num int) string {
	if num == 0 {
		return r.PrettyPrintFirst()
	}

	return at.Yellow("Redirect to: ") + fmt.Sprintf(at.Bold("%s  [%s]"), r.GetRequest(), r.request.Method)
}

func (r WebRequestResult) PrettyPrintNormal(lastStatusCode int) string {
	return fmt.Sprintf("%s%s (%s) %s  [%s] %s", htab, hcont, colorStatus(lastStatusCode), rarrow, r.request.Method, r.GetRequest())
}

func (r WebRequestResult) PrettyPrintLast() string {
	return fmt.Sprintf("%s%s (%s) %s  %s", htab, corner, colorStatus(r.response.StatusCode), rarrow, at.Bold(r.response.Status))
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
	lowerAppName          = strings.ToLower(AppName)
	agentString           = "Go-http-client/2.0 (" + AppName + " Request Analyzer v" + AppVersion + ")"
	rqHeaderDone          = false
	rqCookiesDone         = false
	colorMode             = true
	globalConnSet         ConnectionSetup
	globalRequestBody     string
	globalRequestTemplate = WebRequest{}
	globalHeaderList      []string
	globalCookieLst       []*http.Cookie
	globalHeaderSep       = ":"
	globalCookieSep       = "="
	hcont                 string
	corner                string
	vbar                  string
	htab                  string
	indentHeader          string
	rarrow                string
	screenWidth           int
)
