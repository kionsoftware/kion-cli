package kion

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Console                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

type URLResponse struct {
	Status int    `json:"status"`
	URL    string `json:"data"`
}

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
	// build our query and get response
	url := fmt.Sprintf("%v/api/v1/console-access", host)
	query := map[string]string{}
	data := URLRequest{
		AccountNumber:  car.AccountNumber,
		AccountID:      car.AccountID,
		AccountName:    "",
		AWSIAMRoleName: "",
		AccountTypeID:  0,
		RoleID:         0,
		RoleType:       "",
	}
	resp, _, err := runQuery("POST", url, token, query, data)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
