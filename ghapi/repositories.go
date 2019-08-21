package ghapi

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

// Repository contains repository data
type Repository struct {
	ID               int                   `json:"id"`
	NodeID           string                `json:"node_id"`
	Name             string                `json:"name"`
	FullName         string                `json:"full_name"`
	Owner            User                  `json:"owner"`
	Private          bool                  `json:"private"`
	HTMLURL          string                `json:"html_url"`
	Description      string                `json:"description"`
	Fork             bool                  `json:"fork"`
	URL              string                `json:"url"`
	Homepage         string                `json:"homepage"`
	Language         string                `json:"language"`
	ForksCount       int                   `json:"forks_count"`
	StargazersCount  int                   `json:"stargazers_count"`
	WatchersCount    int                   `json:"watchers_count"`
	Site             int                   `json:"size"`
	DefaultBranch    string                `json:"default_branch"`
	OpenIssuesCount  int                   `json:"open_issues_count"`
	IsTemplate       bool                  `json:"is_template"`
	Topics           []string              `json:"topics"`
	HasIssues        bool                  `json:"has_issues"`
	HasProjects      bool                  `json:"has_projects"`
	HasWiki          bool                  `json:"has_wiki"`
	HasPages         bool                  `json:"has_pages"`
	HasDownloads     bool                  `json:"has_downloads"`
	Archived         bool                  `json:"archived"`
	Disabled         bool                  `json:"disabled"`
	PushedAt         string                `json:"pushed_at"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        string                `json:"updated_at"`
	Permissions      RepositoryPermissions `json:"permissions"`
	SubscribersCount int                   `json:"subscribers_count"`
	License          RepositoryLicense     `json:"license"`
}

// RepositoryPermissions contains permissions data for a repo
type RepositoryPermissions struct {
	Admin bool `json:"admin"`
	Push  bool `json:"push"`
	Pull  bool `json:"pull"`
}

// RepositoryLicense contains license information
type RepositoryLicense struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

// TeamRepositories lists all repos for a team
func (t *Team) TeamRepositories() ([]*Repository, error) {
	var lp ListPages
	var rs []*Repository
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Team Repos for %s, %+v\n", t.Name, lp)
		rsl, llp, err := t.ListRepositories(lp.Next)
		rs = append(rs, rsl...)
		if err != nil {
			return rs, err
		}
		if lp.Next == lp.Last && lp.Last > 0 {
			break
		}
		lp = llp
		if lp.Last == 0 {
			break
		}
	}
	return rs, nil
}

// ListRepositories lists repos for a team
func (t *Team) ListRepositories(page int) ([]*Repository, ListPages, error) {
	var rl []*Repository
	var lp ListPages
	reqURL := "https://api.github.com/teams/" + strconv.Itoa(t.ID) + "/repos"
	if page > 0 {
		reqURL += "?page=" + strconv.Itoa(page)
	}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return rl, lp, err
	}
	req.Header.Set("Accept", "application/vnd.github.hellcat-preview+json")
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return rl, lp, rerr
	}
	links := res.Header.Get("Link")
	lp, err = parseLinks(links)
	if err != nil {
		return rl, lp, err
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return rl, lp, rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return rl, lp, berr
	}
	jerr := json.Unmarshal(bd, &rl)
	if jerr != nil {
		return rl, lp, jerr
	}
	return rl, lp, nil
}

// SaveTeamRepoList saves a repos list to a JSON file
func SaveTeamRepoList(rs []*Repository) error {
	teamRepoFile := path.Join(DataDir, "teamrepos.json")
	os.Remove(teamRepoFile)
	log.SetOutput(os.Stdout)
	log.Printf("Saving team repo list to: %s\n", teamRepoFile)
	jd, jerr := json.Marshal(rs)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(teamRepoFile, jd, 0755)
}
