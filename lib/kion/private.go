package kion

import (
	"encoding/json"
	"fmt"
)

// ConsoleAccessResponse maps to the Kion V1 API response.
type ConsoleAccessCARResponse struct {
	Status            int                `json:"status"`
	ConsoleAccessCARs []ConsoleAccessCAR `json:"data"`
}

// Account maps to the Kion API response for account data.
type ConsoleAccessCAR struct {
	CARName        string    `json:"name"`
	CARID          uint      `json:"id"`
	CARRoleType    string    `json:"role_type"`
	Accounts       []Account `json:"accounts"`
	ConsoleAccess  bool      `json:"console_access"`
	STAKAccess     bool      `json:"short_term_key_access"`
	LTAKAccess     bool      `json:"long_term_key_access"`
	AwsIamRoleName string    `json:"aws_iam_role_name"`
}

func GetConsoleAccessCARS(host string, token string, projID uint) ([]ConsoleAccessCAR, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v1/project/%v/console-access", host, projID)
	query := map[string]string{}
	var data interface{}
	resp, _, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	consoleAccessCarResp := ConsoleAccessCARResponse{}
	err = json.Unmarshal(resp, &consoleAccessCarResp)
	if err != nil {
		return nil, err
	}

	return consoleAccessCarResp.ConsoleAccessCARs, nil
}
