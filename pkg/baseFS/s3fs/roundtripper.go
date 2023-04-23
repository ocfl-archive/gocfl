//go:build exclude

package s3fs

import (
	"fmt"
	"github.com/op/go-logging"
	"net/http"
	"time"
)

/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Copyright 2017 Kazuki Suda.

For the full copyright and license information, please view the LICENSE
file that was distributed with this source code.
*/
/*
Writer replaced by logger
Copyright 2022 Jürgen Enge (Basel)
*/

type requestCanceler interface {
	CancelRequest(*http.Request)
}

// requestInfo keeps track of information about a request/response combination
type requestInfo struct {
	RequestHeaders http.Header
	RequestVerb    string
	RequestURL     string

	ResponseStatus  string
	ResponseHeaders http.Header
	ResponseErr     error

	Duration time.Duration
}

// newRequestInfo creates a new RequestInfo based on an http request
func newRequestInfo(req *http.Request) *requestInfo {
	return &requestInfo{
		RequestURL:     req.URL.String(),
		RequestVerb:    req.Method,
		RequestHeaders: req.Header,
	}
}

// complete adds information about the response to the requestInfo
func (r *requestInfo) complete(response *http.Response, err error) {
	if err != nil {
		r.ResponseErr = err
		return
	}
	r.ResponseStatus = response.Status
	r.ResponseHeaders = response.Header
}

// toCurl returns a string that can be run as a command in a terminal (minus the body)
func (r *requestInfo) toCurl() string {
	headers := ""
	for key, values := range r.RequestHeaders {
		for _, value := range values {
			headers += fmt.Sprintf(` -H %q`, fmt.Sprintf("%s: %s", key, value))
		}
	}

	return fmt.Sprintf("curl -k -v -X%s %s %s", r.RequestVerb, headers, r.RequestURL)
}

// debuggingRoundTripper will display information about the requests passing
// through it based on what is configured
type debuggingRoundTripper struct {
	delegatedRoundTripper http.RoundTripper
	logger                *logging.Logger
	levels                map[debugLevel]bool
}

type debugLevel int

const (
	JustURL debugLevel = iota
	URLTiming
	CurlCommand
	RequestHeaders
	ResponseStatus
	ResponseHeaders
)

func NewDebuggingRoundTripper(rt http.RoundTripper, logger *logging.Logger, levels ...debugLevel) *debuggingRoundTripper {
	drt := &debuggingRoundTripper{
		delegatedRoundTripper: rt,
		logger:                logger,
		levels:                make(map[debugLevel]bool, len(levels)),
	}
	for _, v := range levels {
		drt.levels[v] = true
	}
	return drt
}

func (rt *debuggingRoundTripper) CancelRequest(req *http.Request) {
	if canceler, ok := rt.delegatedRoundTripper.(requestCanceler); ok {
		canceler.CancelRequest(req)
	}
}

func (rt *debuggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqInfo := newRequestInfo(req)

	if rt.levels[JustURL] {
		rt.logger.Debugf("%s %s\n", reqInfo.RequestVerb, reqInfo.RequestURL)
	}
	if rt.levels[CurlCommand] {
		rt.logger.Debugf("%s\n", reqInfo.toCurl())

	}
	if rt.levels[RequestHeaders] {
		str := fmt.Sprintf("Request Headers:\n")
		for key, values := range reqInfo.RequestHeaders {
			for _, value := range values {
				str += fmt.Sprintf("    %s: %s\n", key, value)
			}
		}
		rt.logger.Debug(str)
	}

	startTime := time.Now()
	response, err := rt.delegatedRoundTripper.RoundTrip(req)
	reqInfo.Duration = time.Since(startTime)

	reqInfo.complete(response, err)

	if rt.levels[URLTiming] {
		rt.logger.Debugf("%s %s %s in %d milliseconds\n", reqInfo.RequestVerb, reqInfo.RequestURL, reqInfo.ResponseStatus, reqInfo.Duration.Nanoseconds()/int64(time.Millisecond))
	}
	if rt.levels[ResponseStatus] {
		rt.logger.Debugf("Response Status: %s in %d milliseconds\n", reqInfo.ResponseStatus, reqInfo.Duration.Nanoseconds()/int64(time.Millisecond))
	}
	if rt.levels[ResponseHeaders] {
		str := fmt.Sprintf("Response Headers:\n")
		for key, values := range reqInfo.ResponseHeaders {
			for _, value := range values {
				str += fmt.Sprintf("    %s: %s\n", key, value)
			}
		}
		rt.logger.Debug(str)
	}

	return response, err
}

func (rt *debuggingRoundTripper) WrappedRoundTripper() http.RoundTripper {
	return rt.delegatedRoundTripper
}
