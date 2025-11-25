package structs

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Structs                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Configuration holds the CLI tool values needed to run. The struct maps to
// the applications configured dotfile for persistence between sessions.
type Configuration struct {
	Kion      Kion               `yaml:"kion,omitempty"`
	Favorites []Favorite         `yaml:"favorites,omitempty"`
	Profiles  map[string]Profile `yaml:"profiles,omitempty"`
	Browser   Browser            `yaml:"browser,omitempty"`
}

// Kion holds information about the instance of Kion with which the application
// interfaces with as well as the credentials to do so.
type Kion struct {
	URL              string `yaml:"url,omitempty"`
	APIKey           string `yaml:"api_key,omitempty"`
	Username         string `yaml:"username,omitempty"`
	Password         string `yaml:"password,omitempty"`
	IDMS             string `yaml:"idms_id,omitempty"`
	SamlMetadataFile string `yaml:"saml_metadata_file,omitempty"`
	SamlIssuer       string `yaml:"saml_sp_issuer,omitempty"`
	SamlPrintURL     bool   `yaml:"saml_print_url,omitempty"`
	DisableCache     bool   `yaml:"disable_cache,omitempty"`
	DefaultRegion    string `yaml:"default_region,omitempty"`
	DebugMode        bool   `yaml:"debug_mode,omitempty"`
	QuietMode        bool   `yaml:"quiet_mode,omitempty"`
}

// Favorite holds information about user defined favorites used to quickly
// access desired accounts.
type Favorite struct {
	Name                 string `yaml:"name,omitempty" json:"alias_name"`
	Account              string `yaml:"account,omitempty" json:"account_number"`
	CAR                  string `yaml:"cloud_access_role,omitempty" json:"cloud_access_role_name"`
	AccessType           string `yaml:"access_type,omitempty" json:"access_type"`
	Region               string `yaml:"region,omitempty" json:"account_region"`
	Service              string `yaml:"service,omitempty"`
	FirefoxContainerName string `yaml:"firefox_container_name,omitempty"`
	CloudServiceProvider string `json:"cloud_service_provider"`
	DescriptiveName      string
	Unaliased            bool
}

// Profile holds an alternate configuration for Kion and Favorites.
type Profile struct {
	Kion      Kion       `yaml:"kion,omitempty"`
	Favorites []Favorite `yaml:"favorites,omitempty"`
}

// Browser holds configurations for browser options.
type Browser struct {
	FirefoxContainers bool   `yaml:"firefox_containers,omitempty"`
	CustomBrowserPath string `yaml:"custom_browser_path,omitempty"`
}
