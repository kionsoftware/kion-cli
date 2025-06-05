package structs

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Structs                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Configuration holds the CLI tool values needed to run. The struct maps to
// the applications configured dotfile for persistence between sessions.
type Configuration struct {
	Kion      Kion               `yaml:"kion"`
	Favorites []Favorite         `yaml:"favorites"`
	Profiles  map[string]Profile `yaml:"profiles"`
	Browser   Browser            `yaml:"browser"`
}

// Kion holds information about the instance of Kion with which the application
// interfaces with as well as the credentials to do so.
type Kion struct {
	Url              string `yaml:"url"`
	ApiKey           string `yaml:"api_key"`
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	IDMS             string `yaml:"idms_id"`
	SamlMetadataFile string `yaml:"saml_metadata_file"`
	SamlIssuer       string `yaml:"saml_sp_issuer"`
	SamlPrintUrl     bool   `yaml:"saml_print_url"`
	DisableCache     bool   `yaml:"disable_cache"`
	DefaultRegion    string `yaml:"default_region"`
}

// Favorite holds information about user defined favorites used to quickly
// access desired accounts.
type Favorite struct {
	Name                 string `yaml:"name" json:"alias_name"`
	Account              string `yaml:"account" json:"account_number"`
	CAR                  string `yaml:"cloud_access_role" json:"cloud_access_role_name"`
	AccessType           string `yaml:"access_type" json:"access_type"`
	Region               string `yaml:"region" json:"account_region"`
	Service              string `yaml:"service"`
	FirefoxContainerName string `yaml:"firefox_container_name"`
	CloudServiceProvider string `json:"cloud_service_provider"`
}

// Profile holds an alternate configuration for Kion and Favorites.
type Profile struct {
	Kion      Kion       `yaml:"kion"`
	Favorites []Favorite `yaml:"favorites"`
}

// Browser holds configurations for browser options.
type Browser struct {
	FirefoxContainers bool   `yaml:"firefox_containers"`
	CustomBrowserPath string `yaml:"custom_browser_path"`
}
