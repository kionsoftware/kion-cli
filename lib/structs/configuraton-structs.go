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
	DisableCache     bool   `yaml:"disable_cache"`
}

// Favorite holds information about user defined favorites used to quickly
// access desired accounts.
type Favorite struct {
	Name       string `yaml:"name"`
	Account    string `yaml:"account"`
	CAR        string `yaml:"cloud_access_role"`
	AccessType string `yaml:"access_type"`
	Region     string `yaml:"region"`
}

// Profile holds an alternate configuration for Kion and Favorites.
type Profile struct {
	Kion      Kion       `yaml:"kion"`
	Favorites []Favorite `yaml:"favorites"`
}

type Browser struct {
	FirefoxContainers bool   `yaml:"firefox_containers"`
	CustomBrowserPath string `yaml:"custom_browser_path"`
}
