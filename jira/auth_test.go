package jira

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
)

func TestAuth_RoundTrip(t *testing.T) {
	setup()
	defer teardown()

	req, _ := testClient.NewRequest("GET", "/fake/auth/endpoint", nil)
	testMux.HandleFunc("/fake/auth/endpoint", func(w http.ResponseWriter, r *http.Request) {
		if req == r {
			t.Errorf("Request not cloned %v %v", r, req)
		}

		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			t.Errorf("No Authorization Header set")
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Basic" {
			t.Errorf("Invalid header format %v", authHeader)
		}

		credsStr := parts[1]
		decoded, err := base64.StdEncoding.DecodeString(credsStr)
		if err != nil {
			t.Errorf("Credentials %v not base64 encoded", credsStr)
		}

		creds := strings.Split(string(decoded), ":")
		if len(creds) != 2 || creds[0] != "test-user" || creds[1] != "test-password" {
			t.Errorf("Credentials %v don't match with test-user:test-password", string(decoded))
		}
	})

	_, err := testClient.client.Do(req)
	if err != nil {
		t.Errorf("Error while updating fixed versions %v", err)
	}
}
