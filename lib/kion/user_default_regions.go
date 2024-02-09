package kion

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//                       DEFAULT AWS REGION VARIABLES                         //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// UserConfigResponse maps to the response from the v1/user-config/default-regions API endpoint.
type UserConfigResponse struct {
	Status int `json:"status"`
	Data   struct {
		AwsDefaultCommercialRegion string `json:"aws_default_commercial_region"`
		AwsDefaultGovcloudRegion   string `json:"aws_default_govcloud_region"`
	} `json:"data"`
}

// GetUserDefaultRegions queries the Kion API for the user's default AWS regions.
func GetUserDefaultRegions(host string, token string) (UserConfigResponse, error) {
	// Build the URL for the request
	url := fmt.Sprintf("%s/api/v1/user-config/default-regions", host)

	resp, err := runQuery("GET", url, token, nil, nil)
	if err != nil {
		return UserConfigResponse{}, err
	}

	// Unmarshal response body into the struct
	var defaultRegionResp UserConfigResponse
	err = json.Unmarshal(resp, &defaultRegionResp)
	if err != nil {
		return UserConfigResponse{}, err
	}

	// Check for null values and set defaults
	if defaultRegionResp.Data.AwsDefaultCommercialRegion == "" {
		defaultRegionResp.Data.AwsDefaultCommercialRegion = "us-east-1"
	}
	if defaultRegionResp.Data.AwsDefaultGovcloudRegion == "" {
		defaultRegionResp.Data.AwsDefaultGovcloudRegion = "us-gov-east-1"
	}

	return defaultRegionResp, nil
}
