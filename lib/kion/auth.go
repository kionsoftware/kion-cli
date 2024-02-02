package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Auth                                                                      //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Session maps to the session data returned by Kion after authentication.
type Session struct {
	// ID       int `json:"id"`
	IDMSID   uint
	UserName string
	// UserID   int `json:"user_id"`
	Access struct {
		Expiry string `json:"expiry"`
		Token  string `json:"token"`
		// UserID int    `json:"user_id"`
	} `json:"access"`
	Refresh struct {
		Expiry string `json:"expiry"`
		Token  string `json:"token"`
		// UserID int    `json:"user_id"`
	} `json:"refresh"`
}

// IDMSResponse maps to the Kion API response.
type IDMSResponse struct {
	Status int    `json:"status"`
	IDMSs  []IDMS `json:"data"`
}

// IDMS maps to the Kion API response for configured IDMSs.
type IDMS struct {
	ID         uint   `json:"id"`
	IdmsTypeID uint   `json:"idms_type_id"`
	Name       string `json:"name"`
}

// AuthRequest maps to the required post body when interfacing with the Kion
// API.
type AuthRequest struct {
	IDMSID   uint   `json:"idms"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse maps to the Kion API response.
type AuthResponse struct {
	Status  int     `json:"status"`
	Session Session `json:"data"`
}

// GetIDMSs queries the kion API for all configured IDMS systems with which a
// user can authenticate via username and password.
func GetIDMSs(host string) ([]IDMS, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v2/idms", host)
	query := map[string]string{}
	var data interface{}
	resp, err := runQuery("GET", url, "", query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	idmsResp := IDMSResponse{}
	err = json.Unmarshal(resp, &idmsResp)
	if err != nil {
		return nil, err
	}

	// only pass along idms's that can accept username and password via kion
	idmss := []IDMS{}
	for _, idms := range idmsResp.IDMSs {
		if idms.IdmsTypeID == 1 || idms.IdmsTypeID == 2 {
			idmss = append(idmss, idms)
		}
	}

	return idmss, nil
}

// Authenticate queries the Kion API to authenticate a user via username and
// password.
func Authenticate(host string, idmsID uint, un string, pw string) (Session, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/token", host)
	query := map[string]string{}
	data := AuthRequest{
		IDMSID:   idmsID,
		Username: un,
		Password: pw,
	}
	resp, err := runQuery("POST", url, "", query, data)
	if err != nil {
		return Session{}, err
	}

	// unmarshal response body
	authResp := AuthResponse{}
	err = json.Unmarshal(resp, &authResp)
	if err != nil {
		return Session{}, err
	}

	return authResp.Session, nil
}
