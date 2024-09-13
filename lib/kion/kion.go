package kion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// runQuery performs queries against the Kion API.
func runQuery(method string, url string, token string, query map[string]string, payload interface{}) ([]byte, int, error) {
	// prepare the request body
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	// start our request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, 0, err
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

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// get the body of the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	// handle non 200's
	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, fmt.Errorf("received %v\n %v", resp.StatusCode, string(respBody))
	}

	// return the response
	return respBody, resp.StatusCode, nil
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
	var data interface{}
	resp, _, err := runQuery("GET", url, "", query, data)
	if err != nil {
		return "", err
	}

	// unmarshal response body
	var response struct {
		Status  int    `json:"status"`
		Version string `json:"data"`
	}
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return "", err
	}

	// remove any dev suffixes
	version := strings.Split(response.Version, "-")[0]

	return version, nil
}

// GetSessionDuration returns the AWS session duration configuration Kion uses
// to generate session tokens. If 403 is received, we assume the shortest
// setting of 15 minutes.
func GetSessionDuration(host string, token string) (int, error) {
	url := fmt.Sprintf("%v/api/v3/app-config/aws-access", host)
	query := map[string]string{}
	var data interface{}
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
		Status int `json:"status"`
		Data   struct {
			Duration int `json:"aws_temporary_credentials_duration"`
		} `json:"data"`
	}
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return 0, err
	}

	return response.Data.Duration, nil
}
