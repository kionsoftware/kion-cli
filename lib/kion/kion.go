// Package kion provides functions to interact with the Kion API.
package kion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIRespBody struct {
	Status  int             `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// runQuery performs queries against the Kion API.
func runQuery(method string, url string, token string, query map[string]string, payload any) (APIRespBody, int, error) {
	// prepare our response struct
	apiResp := APIRespBody{}

	// prepare the request body
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return apiResp, 0, err
	}

	// start our request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return apiResp, 0, err
	}

	// append on our parameters to the req.URL.String()
	q := req.URL.Query()
	for key, value := range query {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// add authorization header to the req
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	// identify the source of the request
	req.Header.Add("kion-source", "kion-cli")

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apiResp, 0, err
	}
	defer resp.Body.Close()

	// get the body of the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiResp, 0, err
	}

	err = json.Unmarshal(respBody, &apiResp)
	if err != nil {
		return apiResp, resp.StatusCode, err
	}

	// handle non 200's
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return apiResp, resp.StatusCode, fmt.Errorf("[%v] %v", resp.StatusCode, apiResp.Message)
	}

	// return the response
	return apiResp, resp.StatusCode, nil
}

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Kion Configurations                                                       //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// GetVersion returns the targeted Kion's version number.
func GetVersion(host string) (string, error) {
	url := fmt.Sprintf("%v/api/version", host)
	query := map[string]string{}
	var data any
	resp, _, err := runQuery("GET", url, "", query, data)
	if err != nil {
		return "", err
	}

	// unmarshal response body
	var version string
	err = json.Unmarshal(resp.Data, &version)
	if err != nil {
		return "", err
	}

	// remove any dev suffixes
	version = strings.Split(version, "-")[0]

	return version, nil
}

// GetSessionDuration returns the AWS session duration configuration Kion uses
// to generate session tokens. If 403 is received, we assume the shortest
// setting of 15 minutes.
func GetSessionDuration(host string, token string) (int, error) {
	url := fmt.Sprintf("%v/api/v3/app-config/aws-access", host)
	query := map[string]string{}
	var data any
	resp, status, err := runQuery("GET", url, token, query, data)
	if err != nil {
		if status == 403 {
			return 15, nil
		} else {
			return 0, err
		}
	}

	// unmarshal response body
	var response struct {
		Duration int `json:"aws_temporary_credentials_duration"`
	}
	err = json.Unmarshal(resp.Data, &response)
	if err != nil {
		return 0, err
	}

	return response.Duration, nil
}

// ConvertAccessType converts the access type string between what the API uses
// and the CLI. It converts "console_access" to "web", and vice versa, and
// "short_term_key_access" to "cli" and vice versa. If the access type does
// not match any of these, it returns the original string.
func ConvertAccessType(accessType string) string {
	switch accessType {
	case "console_access":
		return "web"
	case "short_term_key_access":
		return "cli"
	case "web":
		return "console_access"
	case "cli":
		return "short_term_key_access"
	default:
		return accessType
	}
}
