// Package client provides simple retryable http client implementation.
package client

import (
	"net/http"
	"net/url"
	"regexp"
)

var (
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	invalidHeaderErrorRe = regexp.MustCompile(`invalid header`)

	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

type RetryableClient struct {
	client *http.Client
}

func NewRetryableClient() *RetryableClient {
	return &RetryableClient{
		client: &http.Client{},
	}
}

func (r *RetryableClient) Get(targetURL string) (*http.Response, error) {
	resp, err := r.client.Get(targetURL)
	if isRetryable(resp, err) {
		return r.client.Get(targetURL)
	}

	return resp, err
}

func isRetryable(resp *http.Response, err error) bool {
	if err != nil {
		if v, ok := err.(*url.Error); ok {
			if redirectsErrorRe.MatchString(v.Error()) {
				return false
			}

			if schemeErrorRe.MatchString(v.Error()) {
				return false
			}

			if invalidHeaderErrorRe.MatchString(v.Error()) {
				return false
			}

			if notTrustedErrorRe.MatchString(v.Error()) {
				return false
			}
		}

		return true
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}

	return false
}
