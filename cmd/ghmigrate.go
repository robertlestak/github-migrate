package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/umg/devops-github-migrate/ghapi"
)

var (
	migrate  *string
	remove   *string
	org      *string
	team     *string
	dataDir  *string
	token    *string
	pull     *bool
	pullType *string
	users    *bool
	userData *string
	teams    *bool
)

func init() {
	pull = flag.Bool("pull", false, "Pull latest from API")
	pullType = flag.String("type", "all", "Type of data to pull. [collaborators|users|memberships|teams|invitations|repositories|all].")
	migrate = flag.String("migrate", "", "Migrate specified user to SSO")
	remove = flag.String("remove", "", "Remove specified user from org")
	dataDir = flag.String("dir", "", "Directory to store local data. Can be overridden with DATA_DIR env var")
	token = flag.String("token", "", "GitHub token. Can be overridden with GITHUB_TOKEN env var")
	org = flag.String("org", "", "Organization to migrate. Can be overridden with GITHUB_ORG env var")
	team = flag.String("team", "", "Team to operate against. Default to operate against all users")
	users = flag.Bool("users", false, "Print list of users to STDOUT")
	userData = flag.String("data", "login", "Print specific data for a user")
	teams = flag.Bool("teams", false, "Print list of teams to STDOUT")
	flag.Parse()
	if os.Getenv("GITHUB_TOKEN") != "" {
		*token = os.Getenv("GITHUB_TOKEN")
	}
	if os.Getenv("DATA_DIR") != "" {
		*dataDir = os.Getenv("DATA_DIR")
	}
	if os.Getenv("GITHUB_ORG") != "" {
		*org = os.Getenv("GITHUB_ORG")
	}
	ghapi.Org = *org
	ghapi.DataDir = *dataDir
	ghapi.Token = *token
	if *org == "" {
		log.Fatal("org required")
	}
	if *token == "" {
		log.Fatal("token required")
	}
	if *dataDir == "" {
		log.Fatal("data required")
	} else if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		derr := os.MkdirAll(*dataDir, 0755)
		if derr != nil {
			log.Fatal(derr)
		}
	}
}

func pullAll() {
	pullUsers()
	pullMembership()
	pullTeams()
}

func pullData() {
	switch *pullType {
	case "all":
		pullAll()
	case "collaborators":
		pullOutsideCollaborators()
	case "users":
		pullUsers()
	case "memberships":
		pullMembership()
	case "teams":
		pullTeams()
	case "invitations":
		pullInvitations()
	case "repositories":
		pullRepositories()
	}
}

func main() {
	if *pull {
		pullData()
	} else {
		checkAndPull()
	}
	if *migrate != "" {
		u := ghapi.User{
			Login: *migrate,
		}
		err := migrateUser(u)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *remove != "" {
		u := ghapi.User{
			Login: *remove,
		}
		err := removeUser(u)
		if err != nil {
			log.Fatal(err)
		}
	}
	if *teams {
		perr := printTeams()
		if perr != nil {
			log.Fatal(perr)
		}
	}
	if *users {
		if *team != "" {
			perr := printUsersInTeam(*userData, *team)
			if perr != nil {
				log.Fatal(perr)
			}
		} else {
			perr := printUsers(*userData)
			if perr != nil {
				log.Fatal(perr)
			}
		}
	}
}
