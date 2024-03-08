package cmd

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Return values
const (
	OK = iota
	ErrUndef
	ErrNoArg
	ErrCookieJar
	ErrResolve
	ErrGetFlag
	ErrTimeFmt
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
	url       string
	agent     string
	lang      string
	method    string
	authUser  string
	authPass  string
	reqBody   string
	xhdrs     []string
	cookieLst []*http.Cookie
}

func (r *WebRequest) str() string {
	return fmt.Sprintf("%s (%s)", r.url, r.method)
}

type WebRequestResult struct {
	request   http.Request
	response  http.Response
	cookieJar *http.CookieJar
}

func (r *WebRequestResult) str() string {
	return fmt.Sprintf("%s (%s)", r.request.URL.String(), r.response.Status)
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
	agentString  = "Golang Request Checker v" + AppVersion
	rqHeaderDone = false
)
