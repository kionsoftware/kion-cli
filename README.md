Kion CLI
========

Kion CLI is a tool allowing Kion customers to generate short term access keys and federate into the cloud service provider console via the command line.

![kion-cli usage](doc/kion-cli-usage.gif)

Setup
-----

> [!TIP]
> If you are on OSX and use Homebrew you can install with `brew install kionsoftware/tap/kion-cli`.

1. Build the application and place it in your path:

    ```bash
    brew install kionsoftware/tap/kion-cli

    # or

    go build -ldflags "-X main.kionCliVersion=$(cat VERSION.md)" -o kion
    ln -s $(pwd)/kion /usr/local/bin/

    # or

    make install
    ```

2. (optional) For ZSH completion place this file as `_kion` in your ZSH autocomplete path:

    ```sh
    #compdef kion

    _cli_zsh_autocomplete() {
      local -a opts
      local cur
      cur=${words[-1]}
      if [[ "$cur" == "-"* ]]; then
        opts=("${(@f)$(${words[@]:0:#words[@]-1} ${cur} --generate-bash-completion)}")
      else
        opts=("${(@f)$(${words[@]:0:#words[@]-1} --generate-bash-completion)}")
      fi

      if [[ "${opts[1]}" != "" ]]; then
        _describe 'values' opts
      else
        _files
      fi
    }

    compdef _cli_zsh_autocomplete kion
    ```

3. (optional) Create a configuration file in your home directory named `.kion.yml`:

    ```yaml
    ################################################################################
    ##                                                                            ##
    ##  Default Profile                                                           ##
    ##                                                                            ##
    ################################################################################

    kion:
      url: https://mykion.example
      api_key: [api key]
    favorites:
      - name: sandbox
        account: "111122223333"
        cloud_access_role: Admin
        access_type: web               # optional (defaults to cli)
        region: us-gov-west-1          # optional
      - name: prod
        account: "111122224444"
        cloud_access_role: ReadOnly

    ################################################################################
    ##                                                                            ##
    ##  Alternate Profiles                                                        ##
    ##                                                                            ##
    ################################################################################

    profiles:
      dev:
        kion:
          url: https://dev.mykion.example
          api_key: [api key]
          disable_cache: true              # defaults false
        favorites:
          - name: sandbox
            account: 212121212121
            cloud_access_role: Dev
      test:
        kion:
          url: http://test.mykion.example
          api_key: [api key]
    ```

    You can also point Kion CLI to another configuration file by setting the `KION_CONFIG` environment variable to the desired path. See the "Configuration File" section below for all options.

4. Usage examples:

    __Command Line:__

    ```sh
    # open the sandbox AWS console favorited in the config above
    kion fav sandbox

    # generate and print keys for an AWS account
    kion stak --print --account 121212121212 --car Admin

    # start a sub-shell authenticated into an account
    kion stak --account 121212121212 --car Admin

    # start a sub-shell using a wizard to select a target account and Cloud Rule
    kion stak

    # federate into a web console using a wizard to select a target account and Cloud Rule
    # * note that Firefox users will have to approve pop-ups on the first run
    kion console
    ```

    __AWS Profiles:__

    ```config
    [profile one]
    credential_process = /path/to/kion stak --credential-process --account 121212121212 --car DevOps

    [profile two]
    credential_process = /path/to/kion favorite --credential-process MyFavorite
    ```

User Manual
-----------

__Description:__

The Kion CLI allows users to perform common Kion workflows via the command
line. Users can quickly generate short term access keys (stak) or federate
into the cloud service provider web console via configured favorites or by
walking through an account and role selection wizard.

__Commands:__

```text
stak, s            Generate short-term access keys.


favorite, fav, f   Access pre-configured favorites to quickly generate staks or federate
                   into the cloud service provider console depending on the access_type
                   defined in the favorite.

console, con, c    Federate into the cloud service provider console.

run                Run a command with short-term access keys

util               Tools for managing Kion CLI.

help, h            Print usage text.


The following are maintained for compatibility with older Kion utilities:

setenv             Generate short-team access keys and print to standard out.
                   This is effectively 'stak --print'.

savecreds          Store short-term access keys to an AWS credentials profile.
                   This is effectively 'stak --save'.
```

__Files:__

```text
~/.kion.yml       The user configuration file. Defines credentials, target Kion
                  instance, and a list of favorites.
```

__Global Options:__

```text
--endpoint URL, -e URL, --url URL      URL of the Kion instance to interface with.

--user USER, -u USER, --username USER  Username used for authenticating with Kion.

--password PASSWORD, -p PASSWORD       Password used for authenticating with Kion.

--idms IDMS_ID, -i IDMS_ID             IDMS ID with which to authenticate if using
                                       username and password. If only one IDMS is
                                       configured that uses username and password
                                       it is not required to specify its ID.

--saml_metadata_file FILENAME|URL      FILENAME or URL of the identity provider's
                                       XML metadata document.  If a URL, this file
                                       will be downloaded every time the CLI app
                                       is run.  If a local file, this should be an
                                       absolute path to a file on your computer.

--saml_sp_issuer ISSUER                SAML Service Provider issuer value from Kion
                                       for example:
                                       https://mykioninstance.example/api/v1/saml/auth/1

--token TOKEN, -t TOKEN                Token (API or Bearer) used to authenticate.

--disable-cache                        Disable the use of cache for Kion CLI.

--profile PROFILE                      Use the specified PROFILE from the Kion CLI
                                       configuration file. If no profile is specified
                                       the default will be used.

--help, -h                             Print usage text.

--version, -v                          Print the Kion CLI version.
```

__STAK Command:__

```text
OPTIONS

  --print, -p                          Print STAK only. (default: false)

  --account val, --acc val, -a val     Target account number, used to bypass
                                       prompts, must be passed with --car.

  --alias val, --aka val, -l val       Target account alias, used to bypass
                                       prompts, must be passed with --car.

  --car val, --cloud-access-role val,  Target cloud access role, used to bypass
    -c val                             prompts, must be passed with --account
                                       or --alias.

  --region val, -r val                 Specify which region to target.

  --save, -s                           Save short-term keys to an aws credentials
                                       profile. The print flag will supercede this
                                       option.

  --credential-process                 For use with AWS credentials profiles to
                                       setup Kion CLI as a credentials process
                                       subsystem. Returns a json object in the
                                       format needed for the `credential_process`
                                       profile setting.

  --help, -h                           Print usage text.
```

__Favorite Command:__

```text
SUB COMMANDS

  list                                 List all configured favorites. List
                                       accepts a --verbose / -v option to print
                                       additional details.

OPTIONS

  --print, -p                          Print STAK only. Has no effect on
                                       favorites with an "access_type" of
                                       "web". (default: false)

  --credential-process                 For use with AWS credentials profiles to
                                       setup Kion CLI as a credentials process
                                       subsystem. Returns a json object in the
                                       format needed for the `credential_process`
                                       profile setting.

  --help, -h                           Print usage text.
```

__Run Command:__

```text
OPTIONS

  --favorite val, --fav val, -f val    Specify which favorite to run against.

  --account val, -acc val, -a val      Specify which account to target, must be
                                       passed with --car.

  --alias val, --aka val, -l val       Target account alias, used to bypass
                                       prompts, must be passed with --car.

  --car val, -c val                    Specify which Cloud Access Role to use,
                                       must be passed with --account or --alias.

  --region val, -r val                 Specify which region to target.

  --help, -h                           Print usage text.
```

__Util Commands:__

```text
SUB COMMANDS

  flush-cache                          Clear out all cache entries for the Kion CLI.
```

__Environment:__

Environment variables can be set to enable other modalities of Kion CLI usage.
Kion CLI follows standard precedence for defining configurations:

  `Flag > Environment Variable > Configuration File > Default Value`

```text
KION_CONFIG              Path to the Kion CLI configuration file.
                         Defaults to `~/.kion.yml`

KION_URL                 URL of the Kion instance to interact with.

KION_USERNAME            Username used for authenticating with Kion.

KION_PASSWORD            Passwrod used for authenticating with Kion.

KION_IDMS_ID             IDMS ID with which to authenticate if using username and
                         password. If only one IDMS is configured that uses username and
                         password it is not required to specify its ID.

KION_API_KEY             API key used to authenticate. Corresponds to the `--token` flag.

KION_SAML_METADATA_FILE  FILENAME or URL of the identity provider's XML metadata
                         document.  If a URL, this file will be downloaded
                         every time the CLI app is run.  If a local file, this
                         should be an absolute path to a file on your computer.

KION_SAML_SP_ISSUER      The Kion IDMS issuer value, for example
                         https://mykioninstance.example/api/v1/saml/auth/1


The following are maintained for compatibility with older Kion utilities:

CTKEY_USERNAME           Maps to KION_USERNAME.

CTKEY_PASSWORD           Maps to KION_PASSWORD.

CTKEY_APPAPIKEY          Maps to KION_API_KEY
```

__Configuration File:__

```text
KION
----
kion.url                           URL to the target Kion instance.
kion.api_key                       API key used to authenticate.
kion.username                      Username for authentication, to be paried with password.
kion.password                      Password for authentication, to be paired with username.
kion.idms_id                       IDMS ID, if using a custom IDMS in Kion.
kion.saml_metadata_file            SAML metadata file location, URL or path.
kion.saml_sp_issuer                Entity ID for the Kion SAML IDMS.
kion.disable_cache                 Prevents Kion CLI from caching STAK if 'true', defaults 'false'.

FAVORITES
---------
favorites[N].name                  Favorite name, used when calling `kion fav [name]`
favorites[N].account               Account number associated with the favorite.
favorites[N].cloud_access_role     Cloud Access Role used to authenicate with the favorite.
favorites[N].access_type           Favorite access type, 'web' or 'cli', defaults 'cli'.

PROFILES
--------
profiles[NAME].KION                An instance of KION as defined above.
profiles[NAME].FAVORITES           An instance of FAVORITES as defined above.
```

__Caching:__

The Kion CLI has caching enabled by default. The cache is stored in the system keychain and can be disabled by either passing the `--disable-cache` global flag or by setting `kion.disable_cache: true` in the `~/.kion.yml` configuration file. The Kion CLI attempts to receive temporary credential expirations from Kion however if nothing is returned a default credential duration of 15 minutes is set. Cached credentials will be used by default unless:

  - Caching is disabled via the `--disable-cache` global flag
  - Caching is disabled in the `~/.kion.yml` configuration file by setting `disable_cache: true`
  - The credential has less than 5 seconds left and Kion CLI is being used as an AWS credential provider
  - The credential has less than 5 minutes left and Kion CLI is being used to print keys
  - The credential has less than 10 minutes left and Kion CLI is being used to create an AWS configuration profile
  - The credential has less than 5 minutes left and Kion CLI is being used to create an authenticated subshell
  - The credential has less than 5 seconds left and Kion CLI is being used to run an ad hoc command

### Compatibility

Kion-CLI is setup to be a drop in replacement for the older cloudtamer.io
utility `ctkey` where possible. The one exception is the `--iam-role` options
flag is no longer supported and must be replaced with `--cloud-access-role`,
`--car`, or `-c` flag followed by a valid Cloud Access Role name.

```bash
# old ctkey usage to print short-term access keys
ctkey setenv --url=https://YOUR-KION-URL --account=121212121212 --cloud-access-role=admin
# new Kion-CLI example (drop in replacement)
kion setenv --url=https://YOUR-KION-URL --account=121212121212 --cloud-access-role=admin
# new Kion-CLI example (new usage, assumes use of ~/.kion.yml config)
kion stak --print --account 121212121212 --car admin
# new Kion-CLI example (short usage, assumes use of ~/.kion.yml config)
kion s -p -a 121212121212 -c admin

# old ctkey usage to store short-term access keys in an aws configuration profile
ctkey savecreds --url=https://YOUR-KION-URL --account=121212121212 --cloud-access-role=admin
# new Kion-CLI example (drop in replacement)
kion savecreds --url=https://YOUR-KION-URL --account=121212121212 --cloud-access-role=admin
# new Kion-CLI example (new usage, assumes use of ~/.kion.yml config)
kion stak --save --account 121212121212 --car admin
# new Kion-CLI example (short usage, assumes use of ~/.kion.yml config)
kion s -s -a 121212121212 -c admin
```

While you can drop in Kion-CLI directly with old usage it is recommended that
you familiarize yourself with the new methods of access as it is does not
require modification of AWS CLI configuration files.


### SAML Setup

Kion CLI can use the same SAML identity provider as the Kion user interface to
authenticate access to Kion.  This allows for CLI access with no Kion
credentials or API tokens stored to disk.

When SAML is used for authentication, Kion CLI will prompt users to authenticate
via their IDP in a web browser.  Once authenticated through the IDP, the web
page is closed and Kion CLI will use this authenticated session to interact with
the Kion API and generate cloud tokens.

Some extra setup is required to use SAML:

<details>
<summary>Kion Setup</summary>

You must configure Kion to allow proxying a SAML Assertion via the Kion CLI
tool as a supported SAML destination.  This is a supported SAML configuration
but it is not enabled by default.

1. In the Kion app, identify the ID of the SAML IDMS used to log in.  Navigate
   to Users -> Identity Management Systems -> click on the SAML IDMS you use
   to login to Kion.  Locate the ID in the URL of this page.

   For example: `https://mykion.example/portal/idms/##`
2. Using the Kion API, add the Kion CLI tool as an additional SAML destination
   by adding `http://localhost:8400/callback` as a supported destination URL.
   Use the `POST /v3/idms/{id}/destination-url` API.

   For example, if the IDMS ID from the previous step is `2`:

       curl -H "Authorization: Bearer $APIKEY" \
            -X POST \
            -H 'Content-Type: application-json' \
            https://mykion.example/api/v3/idms/2/destination-url \
            -d '{"destination_url": "http://localhost:8400/callback"}'
</details>

<details>
<summary>Kion CLI Configuration</summary>

You must add SAML configuration options to your `~/.kion.yml` file under the
`kion` section:

* `saml_metadata_file` - This is the SAML Metadata XML file provided by your
   IDP.  This should be a path to a file on your computer, or a URL from
   your identity provider.

   Example 1: `/Users/jdoe/.kion/saml-metadata.xml`

   Example 2: `https://dev-XXXXXX.oktapreview.com/app/exkXXXXXXXXXXXXXXXXXXX/sso/saml/metadata`

   To obtain this file:
    * In the Okta Admin UI, this can be found on the SAML application's Sign On
      tab.
    * In the Entra ID UI, this can be found in the SAML application's Endpoints
      section.  Look for the `Federation metadata document`.
* `saml_sp_issuer` - This is the Entity ID for the Kion SAML IDMS.  This can
   be found by navigating to the SAML IDMS in Kion (Users -> Identity Management
   Systems).  Edit the SAML IDMS and copy the `Service Provider Issuer (Entity ID)`
   URL.

   For example: `https://mykion.example/api/v1/saml/auth/2`

</details>

<details>
<summary>Okta Configuration</summary>

Add the Kion CLI URL, `http://localhost:8400/callback` as an [additional
requestable SSO URL](https://support.okta.com/help/s/article/How-to-add-additional-Requestable-SSO-URLs?language=en_US):

1. Login to the Okta administrator UI
2. Browse to the SSO Apps and select the Kion Application you’d like to
   configure.
3. On the **General** tab, scroll down to the **SAML Settings** section and
   click **Edit**
4. Hit Next on Step 1, and get to the Configure SAML section in Step 2 of the
   wizard.
5. Click **Show Advanced Settings**
    1. Under **Other Requestable SSO URLs**, leave the first one as your primary
       FQDN with an Index of `1`.  This will be the normal Kion application
       callback URL such as `https://mykion.example/api/v1/saml/callback`.
    2. Click **+ Add Another** and enter the Kion CLI URL:
        1. URL: `http://localhost:8400/callback`
        2. Index: `2`
6. Click **Next** and then **Finish**.

</details>

<details>
<summary>Entra ID / Azure AD</summary>

Add the Kion CLI URL, `http://localhost:8400/callback` as an [additional
redirect URI](https://learn.microsoft.com/en-us/entra/identity-platform/quickstart-register-app#add-a-redirect-uri):

1. Login to your Entra ID / Azure AD UI
2. Browse to the Entra ID App registrations
3. Find and click on the SAML App Registration for your Kion application
4. Navigate to the Manage -> Authentication section
5. Under the `Redirect URIs` section, click the `Add URI` link to add the Kion
   CLI URL: `http://localhost:8400/callback`
6. Save your changes

</details>

Contributing
------------

1. Ensure you have golang installed and configured.
2. Clone the repository and initialize with `make init`. This will setup the necessary git hooks and other needed tools.
3. Create and populate your `~/.kion.yml` configuration file. See the example at the top of this document.
4. When submitting a PR be sure to note your changes in the `CHANGELOG.md`.
