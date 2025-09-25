package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Projects                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// ProjectResponse maps to the Kion API response.
type ProjectResponse struct {
	Status  int     `json:"status"`
	Project Project `json:"data"`
}

// ProjectsResponse maps to the Kion API response.
type ProjectsResponse struct {
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

// GetProjects queries the Kion API for a list of all projects within the
// application.
func GetProjects(host string, token string) ([]Project, error) {
	// build our query and get response
	url := host + "/api/v3/project"
	query := map[string]string{}
	var data any
	resp, _, err := runQueryWithRetry("GET", url, token, query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	projResp := ProjectsResponse{}
	err = json.Unmarshal(resp, &projResp)
	if err != nil {
		return nil, err
	}

	return projResp.Projects, nil
}

// GetProjectByID returns the project for a given project ID. Note that if a
// user has car access only to a project this will return a 403. To accommodate
// users with minimal permissions test response codes and fallback accordingly
// or use GetProjects which will work but be more verbose.
func GetProjectByID(host string, token string, id uint) (Project, error) {
	// build our query and get response
	url := fmt.Sprintf("%v/api/v3/project/%v", host, id)
	query := map[string]string{}
	var data any
	resp, _, err := runQueryWithRetry("GET", url, token, query, data)
	if err != nil {
		return Project{}, err
	}

	// unmarshal response body
	projResp := ProjectResponse{}
	err = json.Unmarshal(resp, &projResp)
	if err != nil {
		return Project{}, err
	}

	return projResp.Project, nil
}
