package jira

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var (
	testClient *Client
	testMux    *RegexpHandler
	testServer *httptest.Server
)

func setup() {
	testMux = &RegexpHandler{}
	testServer = httptest.NewServer(testMux)
	testBat := BasicAuthTransport{
		Username: "test-user",
		Password: "test-password",
	}
	testClient, _ = NewClient(testBat.Client(), testServer.URL)
}

func teardown() {
	testServer.Close()
}

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(r string, handler http.Handler) {
	pattern := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(r string, handler func(http.ResponseWriter, *http.Request)) {
	pattern := regexp.MustCompile(r)
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func TestJira_UpdateIssue(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/rest/api/2/issue/test-1234", func(w http.ResponseWriter, r *http.Request) {
		if m := r.Method; m != "PUT" {
			t.Errorf("Incorrect HTTP Method. Expected PUT got %v", m)
		}

		if u := r.URL.String(); !strings.HasPrefix("/rest/api/2/issue/test-1234", u) {
			t.Errorf("Incorrect URL. Expected /rest/api/2/issue/test-1234, got %v", u)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	id := "test-1234"
	payload := make(map[string]interface{})
	fields := make(map[string]interface{})
	payload["fields"] = fields
	_, err := testClient.UpdateIssue(id, payload)
	if err != nil {
		t.Errorf("Error on Updating %v", err)
	}
}

func TestJira_UpdateFixVersions(t *testing.T) {
	setup()
	defer teardown()
	testMux.HandleFunc("/rest/api/2/issue/\\w+", func(w http.ResponseWriter, r *http.Request) {
		if m := r.Method; m != "PUT" {
			t.Errorf("Incorrect HTTP Method. Expected PUT got %v", m)
		}

		if u := r.URL.String(); !strings.HasPrefix(u, "/rest/api/2/issue") {
			t.Errorf("Incorrect URL. Expected /rest/api/2/issue, got %v", u)
		}

		w.WriteHeader(http.StatusNoContent)
	})
	tasks := []string{"test-1234", "foo-42"}
	err := testClient.UpdateIssueFixVersions(tasks)
	if err != nil {
		t.Errorf("Error while updating fixed versions %v", err)
	}
}
