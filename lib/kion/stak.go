package kion

import (
	"encoding/json"
	"fmt"
	"time"
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
	Duration        int64  `json:"duration"`
	Expiration      time.Time
}

// STAKRequest maps to the required post body when interfacing with the Kion
// API.
type STAKRequest struct {
	AccountNumber string `json:"account_number"`
	AccountAlias  string `json:"account_alias"`
	CARName       string `json:"cloud_access_role_name"`
}

// GetSTAK queries the Kion API to generate short term access keys. Must pass
// either an account number or an account alias, one can be "".
func GetSTAK(host string, token string, carName string, accNum string, accAlias string) (STAK, error) {
	// only account number or account alias should be provided, use the account
	// number by default
	if accNum != "" && accAlias != "" {
		accAlias = ""
	}

	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/temporary-credentials/cloud-access-role", host)
	query := map[string]string{}
	data := STAKRequest{
		AccountNumber: accNum,
		AccountAlias:  accAlias,
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

	// set the expiration time, buffer by 30 seconds
	duration := stakResp.STAK.Duration
	if duration == 0 {
		duration = 900
	}
	stakResp.STAK.Expiration = time.Now().Add(time.Duration(duration-30) * time.Second)

	return stakResp.STAK, nil
}
