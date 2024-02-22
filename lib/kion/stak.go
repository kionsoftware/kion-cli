package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Short Term Access Keys                                                    //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// STAKResponse maps to the Kion API response.
type STAKResponse struct {
	Status int  `json:"status"`
	STAK   STAK `json:"data"`
}

// STAK maps to the Kion API response for short term access keys.
type STAK struct {
	AccessKey       string `json:"access_key"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

// STAKRequest maps to the required post body when interfacing with the Kion
// API.
type STAKRequest struct {
	AccountNumber string `json:"account_number"`
	CARName       string `json:"cloud_access_role_name"`
}

// GetSTAK queries the Kion API to generate short term access keys.
func GetSTAK(host string, token string, carName string, accNum string) (STAK, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/temporary-credentials/cloud-access-role", host)
	query := map[string]string{}
	data := STAKRequest{
		AccountNumber: accNum,
		CARName:       carName,
	}
	resp, _, err := runQuery("POST", url, token, query, data)
	if err != nil {
		return STAK{}, err
	}

	// unmarshal response body
	stakResp := STAKResponse{}
	err = json.Unmarshal(resp, &stakResp)
	if err != nil {
		return STAK{}, err
	}

	return stakResp.STAK, nil
}
