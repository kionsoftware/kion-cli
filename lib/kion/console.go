package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Console                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// URLResponse maps to the Kion API response.
type URLResponse struct {
	Status int    `json:"status"`
	URL    string `json:"data"`
}

// URLRequest maps to the required post body when interfacing with the Kion
// API.
type URLRequest struct {
	AccountID      uint   `json:"account_id"`
	AccountName    string `json:"account_name"`
	AccountNumber  string `json:"account_number"`
	AWSIAMRoleName string `json:"aws_iam_role_name"`
	AccountTypeID  uint   `json:"account_type_id"`
	RoleID         uint   `json:"role_id"`
	RoleType       string `json:"role_type"`
}

// GetFederationURL queries the Kion API to generate a federation URL.
func GetFederationURL(host string, token string, car CAR) (string, error) {
	// converting cloud access role type to role type
	var roleType string
	switch car.CloudAccessRoleType {
	case "inherited":
		roleType = "ou"
	case "local":
		roleType = "project"
	}

	// build our query and get response
	url := fmt.Sprintf("%v/api/v1/console-access", host)
	query := map[string]string{}
	data := URLRequest{
		AccountID:      car.AccountID,
		AccountName:    car.AccountName,
		AccountNumber:  car.AccountNumber,
		AWSIAMRoleName: car.AwsIamRoleName,
		AccountTypeID:  car.AccountTypeID,
		RoleID:         car.ID,
		RoleType:       roleType,
	}
	resp, _, err := runQuery("POST", url, token, query, data)
	if err != nil {
		return "", err
	}

	// unmarshal response body
	urlResp := URLResponse{}
	err = json.Unmarshal(resp, &urlResp)
	if err != nil {
		return "", err
	}

	return urlResp.URL, nil
}
