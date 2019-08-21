package ghapi

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

// Membership contains a user's org memberships
type Membership struct {
	URL             string       `json:"url"`
	State           string       `json:"state"`
	Role            string       `json:"role"`
	OrganizationURL string       `json:"organization_url"`
	Organization    Organization `json:"organization"`
	User            User         `json:"user"`
}

// Organization contains organization data
type Organization struct {
	Login       string `json:"login"`
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	Location    string `json:"location"`
	Email       string `json:"email"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	HTMLURL     string `json:"html_url"`
	CreatedAt   string `json:"created_at"`
	Type        string `json:"type"`
}

// GetUserMembership lists members in an organization
func (u *User) GetUserMembership() (Membership, error) {
	var ms Membership
	reqURL := "https://api.github.com/orgs/" + Org + "/memberships/" + u.Login
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return ms, err
	}
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return ms, rerr
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return ms, rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return ms, berr
	}
	jerr := json.Unmarshal(bd, &ms)
	if jerr != nil {
		return ms, jerr
	}
	return ms, nil
}

// GetAllMembership gets memberships for all users
func GetAllMembership() ([]Membership, error) {
	var ms []Membership
	var err error
	log.SetOutput(os.Stdout)
	if _, cerr := os.Stat(path.Join(DataDir, "users.json")); os.IsNotExist(cerr) {
		log.Println(path.Join(DataDir, "users.json"), "does not exist")
		return ms, cerr
	}
	ud, uerr := ioutil.ReadFile(path.Join(DataDir, "users.json"))
	if uerr != nil {
		return ms, uerr
	}
	var us []User
	jerr := json.Unmarshal(ud, &us)
	if jerr != nil {
		return ms, jerr
	}
	log.Printf("Getting memberships for all %d members\n", len(us))
	for _, u := range us {
		log.Printf("Getting membership for user: %s", u.Login)
		um, merr := u.GetUserMembership()
		if merr != nil {
			return ms, merr
		}
		ms = append(ms, um)
	}
	return ms, err
}

// SaveMembership saves a membership list to a JSON file
func SaveMembership(ls []Membership) error {
	memberListFile := path.Join(DataDir, "memberships.json")
	os.Remove(memberListFile)
	log.SetOutput(os.Stdout)
	log.Printf("Saving membership list to: %s\n", memberListFile)
	jd, jerr := json.Marshal(ls)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(memberListFile, jd, 0755)
}
