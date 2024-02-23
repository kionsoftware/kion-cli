package kion

import (
	"encoding/json"
	"fmt"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Cloud Access Roles                                                        //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// CARResponse maps to the Kion API response.
type CARResponse struct {
	Status int   `json:"status"`
	CARS   []CAR `json:"data"`
}

// CAR maps to the Kion API response for cloud access roles.
type CAR struct {
	AccountID           uint   `json:"account_id"`
	AccountNumber       string `json:"account_number"`
	AccountType         string `json:"account_type"`
	AccountName         string
	ApplyToAllAccounts  bool   `json:"apply_to_all_accounts"`
	AwsIamPath          string `json:"aws_iam_path"`
	AwsIamRoleName      string `json:"aws_iam_role_name"`
	CloudAccessRoleType string `json:"cloud_access_role_type"`
	CreatedAt           struct {
		Time  time.Time `json:"Time"`
		Valid bool      `json:"Valid"`
	} `json:"created_at"`
	DeletedAt struct {
		Time  time.Time `json:"Time"`
		Valid bool      `json:"Valid"`
	} `json:"deleted_at"`
	FutureAccounts      bool   `json:"future_accounts"`
	ID                  uint   `json:"id"`
	LongTermAccessKeys  bool   `json:"long_term_access_keys"`
	Name                string `json:"name"`
	ProjectID           uint   `json:"project_id"`
	ShortTermAccessKeys bool   `json:"short_term_access_keys"`
	UpdatedAt           struct {
		Time  time.Time `json:"Time"`
		Valid bool      `json:"Valid"`
	} `json:"updated_at"`
	WebAccess bool `json:"web_access"`
}

// GetCARS queries the Kion API for all cloud access roles to which the
// authenticated user has access.
func GetCARS(host string, token string) ([]CAR, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/me/cloud-access-role", host)
	query := map[string]string{}
	var data interface{}
	resp, _, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	carResp := CARResponse{}
	err = json.Unmarshal(resp, &carResp)
	if err != nil {
		return nil, err
	}

	return carResp.CARS, nil
}

// GetCARSOnProject returns all cloud access roles that match a given project and account
func GetCARSOnProject(host string, token string, projID uint, accID uint) ([]CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return nil, err
	}

	// reduce to cars that match project and account, and are not deleted
	var cars []CAR
	for _, car := range allCars {
		if car.ProjectID == projID && car.AccountID == accID && car.DeletedAt.Time.IsZero() {
			cars = append(cars, car)
		}
	}

	return cars, nil
}

// GetCARSOnAccount returns all cloud access roles that match a given account
func GetCARSOnAccount(host string, token string, accID uint) ([]CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return nil, err
	}

	// reduce to cars that match project and account, and are not deleted
	var cars []CAR
	for _, car := range allCars {
		if car.AccountID == accID && car.DeletedAt.Time.IsZero() {
			cars = append(cars, car)
		}
	}

	return cars, nil
}
