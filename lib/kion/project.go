package kion

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Projects                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ProjectResponse maps to the Kion API response.
type ProjectResponse struct {
	Status   int       `json:"status"`
	Projects []Project `json:"data"`
}

// Project maps to the Kion API response for projects.
type Project struct {
	Archived         bool   `json:"archived"`
	AutoPay          bool   `json:"auto_pay"`
	DefaultAwsRegion string `json:"default_aws_region"`
	Description      string `json:"description"`
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	OuID             uint   `json:"ou_id"`
}

// GetProject queries the Kion API for a list of all projects within the application.
func GetProjects(host string, token string) ([]Project, error) {
	// build our query and get response
	url := host + "/api/v3/project"
	query := map[string]string{}
	var data interface{}
	resp, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	projResp := ProjectResponse{}
	err = json.Unmarshal(resp, &projResp)
	if err != nil {
		return nil, err
	}

	return projResp.Projects, nil
}
