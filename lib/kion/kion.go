package kion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	// append on our parameters to the req.URL.String(), only active milestones
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

// getVersion returns the targeted Kions version number.
func GetVersion(host string, token string) (string, error) {
	url := fmt.Sprintf("%v/api/version", host)
	query := map[string]string{}
	var data interface{}
	resp, _, err := runQuery("GET", url, token, query, data)
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

	return response.Version, nil
}
