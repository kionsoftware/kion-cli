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
	AccountTypeID       uint   `json:"account_type_id"`
	AccountName         string `json:"account_name"`
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
// authenticated user has access. Deleted CARs will be excluded.
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

	var cars []CAR
	for _, car := range carResp.CARS {
		if car.DeletedAt.Time.IsZero() {
			cars = append(cars, car)
		}
	}

	return cars, nil
}

// GetCARSOnProject returns all cloud access roles that match a given project and account.
func GetCARSOnProject(host string, token string, projID uint, accID uint) ([]CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return nil, err
	}

	// reduce to cars that match project and account
	var cars []CAR
	for _, car := range allCars {
		if car.ProjectID == projID && car.AccountID == accID {
			cars = append(cars, car)
		}
	}

	return cars, nil
}

// GetCARSOnAccount returns all cloud access roles that match a given account.
func GetCARSOnAccount(host string, token string, accID uint) ([]CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return nil, err
	}

	// reduce to cars that match project and account
	var cars []CAR
	for _, car := range allCars {
		if car.AccountID == accID {
			cars = append(cars, car)
		}
	}

	return cars, nil
}

// GetCARByName returns a car that matches a given name. IMPORTANT: please use
// GetCARByNameAndAccount instead where possible as there are no constraints
// against CARs with duplicate names, this function is kept as a convenience
// and workaround for users on older version of Kion that have limited
// permissions.
func GetCARByName(host string, token string, carName string) (CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return CAR{}, err
	}

	// find our match
	// TODO: build and return a slice of matching CARs, then on all references
	// should be updated to handle the slice and prompt users for selection or
	// test all for success silently
	for _, car := range allCars {
		if car.Name == carName {
			return car, nil
		}
	}

	return CAR{}, fmt.Errorf("unable to find car %v", carName)
}

// GetCARByNameAndAccount returns a car that matches by name and account number.
func GetCARByNameAndAccount(host string, token string, carName string, accountNumber string) (CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return CAR{}, err
	}

	// find our match
	for _, car := range allCars {
		if car.Name == carName && car.AccountNumber == accountNumber {
			return car, nil
		}
	}

	return CAR{}, fmt.Errorf("unable to find car %v", carName)
}

// GetAllCARsByName returns a slice of cars that matches a given name.
func GetAllCARsByName(host string, token string, carName string) ([]CAR, error) {
	allCars, err := GetCARS(host, token)
	if err != nil {
		return nil, err
	}

	// find our matches
	var cars []CAR
	for _, car := range allCars {
		if car.Name == carName {
			account, _, err := GetAccount(host, token, car.AccountNumber)
			if err != nil {
				// TODO: this may not be what we want to do here, kept as info level log
				// fmt.Println("  unable to lookup an associated account:", car.AccountNumber)
				continue
			}
			car.AccountName = account.Name
			car.AccountTypeID = account.TypeID
			cars = append(cars, car)
		}
	}

	// return our slice of cars
	if len(cars) > 0 {
		return cars, nil
	}

	return nil, fmt.Errorf("unable to find car %v", carName)
}
