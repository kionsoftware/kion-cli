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

// GetIDMSs queries the Kion API for all configured IDMS systems with which a
// user can authenticate via username and password.
func GetIDMSs(host string) ([]IDMS, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v2/idms", host)
	query := map[string]string{}
	var data any
	resp, _, err := runQuery("GET", url, "", query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	var idmss []IDMS
	err = json.Unmarshal(resp.Data, &idmss)
	if err != nil {
		return nil, err
	}

	// only pass along idms's that can accept username and password via kion
	unpwIdmss := []IDMS{}
	for _, idms := range idmss {
		if idms.IdmsTypeID == 1 || idms.IdmsTypeID == 2 {
			unpwIdmss = append(unpwIdmss, idms)
		}
	}

	return unpwIdmss, nil
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
	resp, _, err := runQuery("POST", url, "", query, data)
	if err != nil {
		return Session{}, err
	}

	// unmarshal response body
	var session Session
	err = json.Unmarshal(resp.Data, &session)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}
