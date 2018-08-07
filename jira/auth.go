package jira

import (
	"net/http"
)

// BasicAuthTransport is the Credentials struct
type BasicAuthTransport struct {
	Username string
	Password string
}

// RoundTrip Appends the Credentials after cloning the request
func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := cloneRequest(req)
	reqClone.SetBasicAuth(bat.Username, bat.Password)
	return http.DefaultTransport.RoundTrip(reqClone)
}

// Client Returns the client of the Transport
func (bat *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: bat}
}

func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
