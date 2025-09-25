package kion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Helpers                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// runQuery performs queries against the Kion API.
func runQuery(method string, url string, token string, query map[string]string, payload any) ([]byte, int, error) {
	// ...existing code...
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

	// identify the source of the request
	req.Header.Add("kion-source", "kion-cli")

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

// runQueryWithRetry wraps runQuery with intelligent retry logic.
func runQueryWithRetry(method string, url string, token string, query map[string]string, payload any) ([]byte, int, error) {
	const maxRetries = 3
	baseDelay := 500 * time.Millisecond

	var lastErr error
	var lastStatus int

	for attempt := range maxRetries {
		resp, status, err := runQuery(method, url, token, query, payload)

		// Success case
		if err == nil {
			return resp, status, nil
		}

		lastErr = err
		lastStatus = status

		// Don't retry on non-retryable errors
		if !isRetryableError(status) {
			return resp, status, err
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries-1 {
			// Exponential backoff with jitter
			delay := time.Duration(attempt+1) * baseDelay
			time.Sleep(delay)
		}
	}

	return nil, lastStatus, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError determines if an error should be retried based on status
// code and error type.
func isRetryableError(statusCode int) bool {
	// Network-level errors should be retried
	if statusCode == 0 {
		return true
	}

	// Retry on these HTTP status codes
	switch statusCode {
	case 408: // Request Timeout
		return true
	case 429: // Too Many Requests (rate limiting)
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	default:
		return false
	}
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
	resp, _, err := runQueryWithRetry("GET", url, "", query, data)
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
	var data any
	resp, status, err := runQueryWithRetry("GET", url, token, query, data)
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
