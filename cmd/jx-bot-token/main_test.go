package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestOptions_findErrors(t *testing.T) {

	type fields struct {
		gitProvider  string
		responseCode int
	}

	tests := []struct {
		name string
		fields
		want []string
	}{
		{name: "unknown_git_provider", fields: fields{gitProvider: "https://foo.com"}, want: []string{"verifying bot token not yet supported for git provider https://foo.com"}},
		{name: "github_ok", fields: fields{gitProvider: "https://github.com", responseCode: 200}, want: nil},
		{name: "github_error", fields: fields{gitProvider: "https://github.com", responseCode: 401}, want: []string{"failed to verify bot account with https://api.github.com/user, response code 401"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Start a local HTTP server
			client := NewTestClient(func(req *http.Request) *http.Response {
				// Test request parameters
				return &http.Response{
					StatusCode: tt.responseCode,
					// Send response to be tested
					Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
					// Must be set to non-nil value or it panics
					Header: make(http.Header),
				}
			})

			os.Setenv(gitProvider, tt.gitProvider)

			o := Options{
				httpClient: client,
			}

			if got := o.findErrors(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
