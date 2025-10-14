package commands

import (
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/kionsoftware/kion-cli/lib/helper"
	"github.com/kionsoftware/kion-cli/lib/kion"
	samlTypes "github.com/russellhaering/gosaml2/types"
	"github.com/urfave/cli/v2"
)

// authUNPW prompts for any missing credentials then auths the users against
// Kion, stores the session data, and sets the context token.
func (c *Cmd) authUNPW(cCtx *cli.Context) error {
	var err error
	un := c.config.Kion.Username
	pw := c.config.Kion.Password
	idmsID := cCtx.Uint("idms")

	// prompt idms if needed
	if idmsID == 0 {
		idmss, err := kion.GetIDMSs(c.config.Kion.URL)
		if err != nil {
			return err
		}
		iNames, iMap := helper.MapIDMSs(idmss)
		if len(iNames) > 1 {
			idms, err := helper.PromptSelect("Select Login IDMS:", iNames)
			if err != nil {
				return err
			}
			idmsID = iMap[idms].ID
		} else {
			idmsID = iMap[iNames[0]].ID
		}
	}

	// prompt username if needed
	if un == "" {
		un, err = helper.PromptInput("Username:")
		if err != nil {
			return err
		}
	}

	// prompt password if needed
	pwFoundInCache := false
	if pw == "" {
		// Check password cache
		pw, pwFoundInCache, err = c.cache.GetPassword(c.config.Kion.URL, idmsID, un)
		if err != nil {
			return err
		}

		if !pwFoundInCache {
			pw, err = helper.PromptPassword("Password:")
			if err != nil {
				return err
			}
		}
	}

	// auth and capture our session
	session, err := kion.Authenticate(c.config.Kion.URL, idmsID, un, pw)
	if err != nil {
		// Unfortunately, the remote auth endpoint doesn't provide an easy way
		// of determining if an auth error was the cause of failure (it returns
		// an HTTP 400 with a body that contains a message about authentication
		// failues). Conservatively clear out any cached password when
		// Authenticate() fails
		if pwFoundInCache {
			err := c.cache.SetPassword(c.config.Kion.URL, idmsID, un, "")
			if err != nil {
				// We're already handling another error, logging
				// is the best we can do
				color.Red("Failed to clear password from cache, %v", err)
			}
		}
		return err
	}
	session.IDMSID = idmsID
	session.UserName = un
	err = c.cache.SetSession(session)
	if err != nil {
		return err
	}

	// if auth succeeded, cache the password
	err = c.cache.SetPassword(c.config.Kion.URL, idmsID, un, pw)
	if err != nil {
		return err
	}

	// set our token in the config
	c.config.Kion.APIKey = session.Access.Token
	return nil
}

// authSAML directs the user to authenticate via SAML in a web browser.
// The SAML assertion is posted to this app which is forwarded to Kion and
// exchanged for the context token.
func (c *Cmd) authSAML(cCtx *cli.Context) error {
	var err error
	samlMetadataFile := c.config.Kion.SamlMetadataFile
	samlServiceProviderIssuer := c.config.Kion.SamlIssuer

	// prompt metadata url if needed
	if samlMetadataFile == "" {
		samlMetadataFile, err = helper.PromptInput("SAML Metadata URL:")
		if err != nil {
			return err
		}
	}

	// prompt issuer if needed
	if samlServiceProviderIssuer == "" {
		samlServiceProviderIssuer, err = helper.PromptInput("SAML Service Provider Issuer:")
		if err != nil {
			return err
		}
	}

	var samlMetadata *samlTypes.EntityDescriptor
	if strings.HasPrefix(samlMetadataFile, "http") {
		samlMetadata, err = kion.DownloadSAMLMetadata(samlMetadataFile)
		if err != nil {
			return err
		}
	} else {
		samlMetadata, err = kion.ReadSAMLMetadataFile(samlMetadataFile)
		if err != nil {
			return err
		}
	}
	var authData *kion.AuthData

	// we only need to check for existence - the value is irrelevant
	if cCtx.App.Metadata["useOldSAML"] == true {
		authData, err = kion.AuthenticateSAMLOld(
			c.config.Kion.URL,
			samlMetadata,
			samlServiceProviderIssuer,
			c.config.Kion.SamlPrintURL,
		)
		if err != nil {
			return err
		}
	} else {
		authData, err = kion.AuthenticateSAML(
			c.config.Kion.URL,
			samlMetadata,
			samlServiceProviderIssuer,
			c.config.Kion.SamlPrintURL,
		)
		if err != nil {
			return err
		}
	}

	// cache the session for 9.5 minutes, tokens are valid for 10 minutes
	timeFormat := "2006-01-02T15:04:05-0700"
	session := kion.Session{
		Access: struct {
			Expiry string `json:"expiry"`
			Token  string `json:"token"`
		}{
			Token:  authData.AuthToken,
			Expiry: time.Now().Add(570 * time.Second).Format(timeFormat),
		},
	}
	err = c.cache.SetSession(session)
	if err != nil {
		return err
	}

	// set our token in the config
	c.config.Kion.APIKey = authData.AuthToken
	return nil
}

// setAuthToken sets the token to be used for querying the Kion API. If not
// passed to the tool as an argument, set in the env, or present in the
// configuration dotfile it will prompt the users to authenticate. Auth methods
// are prioritized as follows: api/bearer token -> username/password -> saml.
// If flags are set for multiple methods the highest priority method will be
// used.
func (c *Cmd) setAuthToken(cCtx *cli.Context) error {
	if c.config.Kion.APIKey == "" {
		// if we still have an active session use it
		session, found, err := c.cache.GetSession()
		if err != nil {
			return err
		}
		if found && session.Access.Expiry != "" {
			timeFormat := "2006-01-02T15:04:05-0700"
			now := time.Now()
			expiration, err := time.Parse(timeFormat, session.Access.Expiry)
			if err != nil {
				return err
			}
			if expiration.After(now) {
				// TODO: test token is good with an endpoint that is accessible to all
				// user permission levels, if you get a 401 then assume token is bad
				// due to caching a cred when a users password expired, and flush the
				// cache instead...
				c.config.Kion.APIKey = session.Access.Token
				return nil
			}

			// TODO: uncomment when / if the application supports refresh tokens

			// // see if we can use the refresh token
			// refreshExp, err := time.Parse(timeFormat, session.Refresh.Expiry)
			// if err != nil {
			// 	return err
			// }

			// if refreshExp.After(now) {
			// 	un := session.UserName
			// 	idmsId := session.IDMSID
			// 	session, err = kion.Authenticate(c.config.Kion.Url, idmsId, un, session.Refresh.Token)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	session.UserName = un
			// 	session.IDMSID = idmsId
			// 	err = c.cache.SetSession(session)
			// 	if err != nil {
			// 		return err
			// 	}

			//  c.config.Kion.ApiKey = session.Access.Token
			// 	return nil
			// }
		}

		// check un / pw were set via flags and infer auth method
		if c.config.Kion.Username != "" || c.config.Kion.Password != "" {
			err := c.authUNPW(cCtx)
			return err
		}

		// check if saml auth flags set and auth with saml if so
		if c.config.Kion.SamlMetadataFile != "" && c.config.Kion.SamlIssuer != "" {
			err := c.authSAML(cCtx)
			return err
		}

		// if no token or session found, prompt for desired auth method
		methods := []string{
			"API Key",
			"Password",
			"SAML",
		}
		authMethod, err := helper.PromptSelect("How would you like to authenticate", methods)
		if err != nil {
			return err
		}

		// handle chosen auth method
		switch authMethod {
		case "API Key":
			apiKey, err := helper.PromptPassword("API Key:")
			if err != nil {
				return err
			}
			c.config.Kion.APIKey = apiKey
		case "Password":
			err := c.authUNPW(cCtx)
			if err != nil {
				return err
			}
		case "SAML":
			err := c.authSAML(cCtx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
