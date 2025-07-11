package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Accounts                                                                  //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// AccountResponse maps to the Kion API response.
type AccountResponse struct {
	Status  int     `json:"status"`
	Account Account `json:"data"`
}

// AccountsResponse maps to the Kion API response.
type AccountsResponse struct {
	Status   int       `json:"status"`
	Accounts []Account `json:"data"`
}

// Account maps to the Kion API response for account data.
type Account struct {
	Email                     string `json:"account_email"`
	Name                      string `json:"account_name"`
	Alias                     string `json:"account_alias"`
	Number                    string `json:"account_number"`
	TypeID                    uint   `json:"account_type_id"`
	ID                        uint   `json:"id"`
	IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
	LinkedAccountNumber       string `json:"linked_account_number"`
	LinkedRole                string `json:"linked_role"`
	PayerID                   uint   `json:"payer_id"`
	ProjectID                 uint   `json:"project_id"`
	SkipAccessChecking        bool   `json:"skip_access_checking"`
	UseOrgAccountInfo         bool   `json:"use_org_account_info"`
}

// GetAccountsOnProject returns a list of Accounts associated with a given Kion
// project.
func GetAccountsOnProject(host string, token string, id uint) ([]Account, int, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/project/%v/accounts", host, id)
	query := map[string]string{}
	var data any
	resp, statusCode, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, statusCode, err
	}

	// unmarshal response body
	accResp := AccountsResponse{}
	jsonErr := json.Unmarshal(resp, &accResp)
	if jsonErr != nil {
		return nil, 0, err
	}

	return accResp.Accounts, accResp.Status, nil
}

// GetAccount returns an account by the given account number.
func GetAccount(host string, token string, accountNum string) (*Account, int, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/account/by-account-number/%v", host, accountNum)
	query := map[string]string{}
	var data any
	resp, statusCode, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, statusCode, err
	}

	// unmarshal response body
	accResp := AccountResponse{}
	jsonErr := json.Unmarshal(resp, &accResp)
	if jsonErr != nil {
		return nil, 0, err
	}

	return &accResp.Account, accResp.Status, nil
}
