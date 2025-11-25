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
	var data any
	resp, _, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return nil, err
	}

	// unmarshal response body
	var projects []Project
	err = json.Unmarshal(resp.Data, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
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
	resp, _, err := runQuery("GET", url, token, query, data)
	if err != nil {
		return Project{}, err
	}

	// unmarshal response body
	var project Project
	err = json.Unmarshal(resp.Data, &project)
	if err != nil {
		return Project{}, err
	}

	return project, nil
}
