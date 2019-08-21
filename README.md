# GitHub Migrate

This tool facilitates migrating users in GitHub from one organization to another, or from a local-auth organization to an SSO-enabled organization.

## Building

`make` will compile application into `dist` directory.

## Configuration

All configuration parameters can either be provided as command line flags (see `ghmigrate -h` for all flags), or as environment variables (see `.env-sample`).

The user running the application must be an owner of the organization GitHub account.

## High Level Migration Process

To complete a migration, the following process must be followed:

- Pull users and associations from current organization
- Convert organization to an SSO-enabled organization, configure with AzureAD. All users added after this point will be SSO-enabled accounts.
- Remove all current users from organization
- Re-add users to organization. GitHub will enforce SSO when the user attempts to log into organization

## Usage

Ensure that your `GITHUB_TOKEN` exists either in your environment, in your `.env` file, or that you are passing the token in with the `-token` flag.

### Pull user data

`ghmigrate -org <ORG> -dir <DATA_DIR> -pull`

Will download all user data and organizational mappings to `json` files in the data directory.

### List all organization users

`ghmigrate -dir <DATA_DIR> -users`

Will output all users in the organization, one user per line.

### List all team users

`ghmigrate -dir <DATA_DIR> -users -team <TEAM_SLUG>`

Will output all users in the team, one user per line.

### List all teams

`ghmigrate -dir <DATA_DIR> -teams`

Will output all teams in the organization, one user per line.

### List emails for users in team

`ghmigrate -users -team devops -data email`

Will output emails for all users in devops team.

### Migrate user

`ghmigrate -dir <DATA_DIR> -org <ORG> -migrate <USERNAME>`

Will remove the `username` from the organization and then re-add the user, assuming the account has been converted to SSO-enabled.

As the migration command accepts a single user as the input, this should be scripted in conjunction with either the users listing command or the teams listing command to migrate large blocks of users at a time.

*NOTE*: The user will be re-added back to the team(s) they were previously a member of, with all existing rights / access.

### Example Usage

The following outlines a complete organization migration, with some additional examples of individual user and team migrations.

````
# Set config values. Can also be done in .env file or CLI flags.
export GITHUB_TOKEN=abcd1234
export DATA_DIR=data
export GITHUB_ORG=umg

# Pull all users and mappings to the DATA_DIR
ghmigrate -pull

# List all users in organization
ghmigrate -users

# List all users in `metadata-services` team
ghmigrate -users -team metadata-services

# Migrate individual user `lestakr`
ghmigrate -migrate lestakr

# Migrate all users in team `metadata-services`
for user in $(ghmigrate -users -team metadata-services); do
  ghmigrate -migrate "$user"
done

````

## GitHub API Stability

This application makes use of both `alpha` and `beta` API endpoints in GitHub.

If issues are experienced, confirm API stability / availability with GitHub.

API documentation is available here: https://developer.github.com/v3/
