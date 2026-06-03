package commands

import (
	"fmt"
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
			idms, err := helper.PromptSelect("Select Login IDMS:", "Select your IDMS from the list below.", iNames)
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

	// Validate Kion URL is configured
	if c.config.Kion.URL == "" {
		return fmt.Errorf("the Kion URL is not configured; please set 'url' in your configuration file or use the --url flag")
	}

	// prompt metadata url if needed
	if samlMetadataFile == "" {
		samlMetadataFile, err = helper.PromptInput("SAML Metadata URL or File Path:")
		if err != nil {
			return err
		}
		if samlMetadataFile == "" {
			return fmt.Errorf("SAML Metadata URL/File is required for SAML authentication")
		}
	}

	// prompt issuer if needed
	if samlServiceProviderIssuer == "" {
		samlServiceProviderIssuer, err = helper.PromptInput("SAML Service Provider Issuer:")
		if err != nil {
			return err
		}
		if samlServiceProviderIssuer == "" {
			return fmt.Errorf("SAML Service Provider Issuer is required for SAML authentication")
		}
	}

	var samlMetadata *samlTypes.EntityDescriptor
	if strings.HasPrefix(samlMetadataFile, "http") {
		samlMetadata, err = kion.DownloadSAMLMetadata(samlMetadataFile)
		if err != nil {
			return fmt.Errorf("failed to download SAML metadata: %w", err)
		}
	} else {
		samlMetadata, err = kion.ReadSAMLMetadataFile(samlMetadataFile)
		if err != nil {
			return fmt.Errorf("failed to read SAML metadata file: %w", err)
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
	if authData.RefreshToken != "" && !authData.RefreshExpiry.IsZero() {
		session.Refresh.Token = authData.RefreshToken
		session.Refresh.Expiry = authData.RefreshExpiry.Format(timeFormat)
	}
	err = c.cache.SetSession(session)
	if err != nil {
		return err
	}

	// set our token in the config
	c.config.Kion.APIKey = authData.AuthToken
	return nil
}

// sessionTimeFormat is the layout used for Access.Expiry and Refresh.Expiry
// in the cached session.
const sessionTimeFormat = "2006-01-02T15:04:05-0700"

// tryRefreshSession attempts to refresh an expired (or near-expired) cached
// session's access token using its refresh token. Returns the refreshed
// session on success and a boolean indicating whether a usable session was
// produced. Errors are non-fatal — callers fall through to fresh auth when
// refresh isn't possible.
func (c *Cmd) tryRefreshSession(session kion.Session) (kion.Session, bool) {
	if session.Refresh.Token == "" || session.Refresh.Expiry == "" {
		return kion.Session{}, false
	}
	refreshExp, err := time.Parse(sessionTimeFormat, session.Refresh.Expiry)
	if err != nil || !refreshExp.After(time.Now()) {
		return kion.Session{}, false
	}
	refreshed, err := kion.RefreshSession(c.config.Kion.URL, session.Refresh.Token)
	if err != nil || refreshed.Access.Token == "" {
		return kion.Session{}, false
	}
	// the refresh endpoint returns only a new access token — carry the
	// existing identity and refresh token forward.
	refreshed.UserName = session.UserName
	refreshed.IDMSID = session.IDMSID
	refreshed.Refresh = session.Refresh
	if err := c.cache.SetSession(refreshed); err != nil {
		return kion.Session{}, false
	}
	return refreshed, true
}

// freshAPIKey returns a bearer token that is guaranteed to be valid for at
// least the next 30 seconds. It re-checks the cached session's expiry at
// every call, refreshing transparently when the access token is close to
// expiring. This is the correct accessor at any API-call boundary — the
// initial token set by setAuthToken may have expired by the time we get
// around to using it (e.g. user paused at an interactive selection prompt).
//
// If the APIKey came from a flag/env/config (not a cached session) it is
// returned as-is — externally-supplied API keys do not have refresh tokens.
func (c *Cmd) freshAPIKey() (string, error) {
	if c.config.Kion.APIKey == "" {
		return "", fmt.Errorf("no API key available; authenticate first")
	}

	session, found, err := c.cache.GetSession()
	if err != nil {
		return "", err
	}
	// no cached session means the APIKey was set externally (flag/env/config)
	// or via an API-key prompt — nothing to refresh.
	if !found || session.Access.Token == "" {
		return c.config.Kion.APIKey, nil
	}
	// guard against a cached session that belongs to a different identity
	// than the currently-active APIKey (rare, but possible if a flag
	// overrides config mid-session).
	if session.Access.Token != c.config.Kion.APIKey {
		return c.config.Kion.APIKey, nil
	}

	expiration, err := time.Parse(sessionTimeFormat, session.Access.Expiry)
	if err != nil {
		// unparseable expiry — fall back to returning what we have rather
		// than failing the command; the API call will surface a 401 if bad.
		return c.config.Kion.APIKey, nil
	}
	if expiration.After(time.Now().Add(30 * time.Second)) {
		return c.config.Kion.APIKey, nil
	}

	// access token has expired (or will within 30s) — try refresh.
	refreshed, ok := c.tryRefreshSession(session)
	if !ok {
		return "", fmt.Errorf("session expired and could not be refreshed; please re-run to authenticate")
	}
	c.config.Kion.APIKey = refreshed.Access.Token
	return refreshed.Access.Token, nil
}

// setAuthToken ensures an API key is set on the Cmd. It is invoked once at
// the start of an authed command. Order of precedence: api/bearer token ->
// cached session (with refresh) -> username/password -> saml. If no method
// is configured the user is prompted to choose one.
//
// Tokens may expire between when setAuthToken runs and when they are
// actually used (interactive prompts can take a while). Always call
// freshAPIKey at API-call sites to pick up a refreshed token.
func (c *Cmd) setAuthToken(cCtx *cli.Context) error {
	if c.config.Kion.APIKey != "" {
		return nil
	}

	// if we still have an active or refreshable session use it
	session, found, err := c.cache.GetSession()
	if err != nil {
		return err
	}
	if found && session.Access.Expiry != "" {
		expiration, err := time.Parse(sessionTimeFormat, session.Access.Expiry)
		if err != nil {
			return err
		}
		// a small buffer here avoids handing the rest of the command a
		// near-dead token; freshAPIKey handles the rest at call sites.
		if expiration.After(time.Now().Add(30 * time.Second)) {
			// TODO: test token is good with an endpoint that is accessible to all
			// user permission levels, if you get a 401 then assume token is bad
			// due to caching a cred when a users password expired, and flush the
			// cache instead...
			c.config.Kion.APIKey = session.Access.Token
			return nil
		}
		if refreshed, ok := c.tryRefreshSession(session); ok {
			c.config.Kion.APIKey = refreshed.Access.Token
			return nil
		}
		// refresh failed — fall through to the normal auth path below.
	}

	// check un / pw were set via flags and infer auth method
	if c.config.Kion.Username != "" || c.config.Kion.Password != "" {
		return c.authUNPW(cCtx)
	}

	// check if saml auth flags set and auth with saml if so
	if c.config.Kion.SamlMetadataFile != "" && c.config.Kion.SamlIssuer != "" {
		return c.authSAML(cCtx)
	}

	// if no token or session found, prompt for desired auth method
	methods := []string{
		"API Key",
		"Password",
		"SAML",
	}
	authMethod, err := helper.PromptSelect("How would you like to authenticate?", "Choose your preferred authentication method.", methods)
	if err != nil {
		return err
	}

	switch authMethod {
	case "API Key":
		apiKey, err := helper.PromptPassword("API Key:")
		if err != nil {
			return err
		}
		c.config.Kion.APIKey = apiKey
	case "Password":
		return c.authUNPW(cCtx)
	case "SAML":
		return c.authSAML(cCtx)
	}
	return nil
}
