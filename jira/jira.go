package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// Client JIRA consumer struct
type Client struct {
	client  *http.Client
	baseURL *url.URL
}

// NewClient Creates a new JIRA Client
func NewClient(httpClient *http.Client, baseURL string) (*Client, error) {
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	c := &Client{client: httpClient, baseURL: parsedBaseURL}
	return c, nil
}

// NewRequest creates a new http request with a specified **method**. The final url is baseURL + endpoint
func (c *Client) NewRequest(method string, endpoint string, body interface{}) (*http.Request, error) {
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	// Relative URLs should be specified without a preceding slash since baseURL will have the trailing slash
	parsedEndpoint.Path = strings.TrimLeft(parsedEndpoint.Path, "/")

	url := c.baseURL.ResolveReference(parsedEndpoint)

	// Set request body
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}
	// We are only working with JSON
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// UpdateIssue is a Generic update issue based on a JSON like paylod
func (c *Client) UpdateIssue(issueID string, data map[string]interface{}) (*http.Response, error) {
	endpoint := fmt.Sprintf("/rest/api/2/issue/%s", issueID)
	req, err := c.NewRequest("PUT", endpoint, data)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// UpdateIssueFixVersions Updates fixed versions of an array of tasks
func (c *Client) UpdateIssueFixVersions(tasks []string) error {
	fv := viper.GetString("jira.draft-version")
	for _, task := range tasks {
		_, err := c.UpdateIssue(task, NewFixedVersionBody(fv))
		if err != nil {
			return err
		}
		fmt.Printf("Changing fixed version for task %v to %v\n", task, fv)
	}
	return nil
}

// NewActionBody creates JSON payload for a new action
func NewActionBody(action string, updateOp map[string]interface{}) map[string]interface{} {
	root := make(map[string]interface{})
	root[action] = make(map[string]interface{})
	root[action] = updateOp
	return root
}

// NewUpdateOp creates JSON payload for a new update operation
func NewUpdateOp(op string, opData []map[string]interface{}) map[string]interface{} {
	updateOp := make(map[string]interface{})
	updateOp[op] = opData

	return NewActionBody("update", updateOp)
}

// NewFixedVersionBody Creates the fixed body version payload
func NewFixedVersionBody(v string) map[string]interface{} {
	var setOps [1]map[string]string
	nameOp := make(map[string]string)
	nameOp["name"] = v
	setOps[0] = nameOp

	var fixedVersions []map[string]interface{}
	fixedVersions = make([]map[string]interface{}, 1)

	fv := make(map[string]interface{})
	fv["set"] = setOps
	fixedVersions[0] = fv

	return NewUpdateOp("fixVersions", fixedVersions)
}
