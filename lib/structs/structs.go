package structs

import "github.com/kionsoftware/kion-cli/lib/kion"

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Structs                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// Configuration holds the CLI tool values needed to run. The struct maps to
// the applications configured dotfile for persistence between sessions.
type Configuration struct {
	Kion      Kion         `yaml:"kion"`
	Session   kion.Session `yaml:"session"`
	Favorites []Favorite   `yaml:"favorites"`
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
