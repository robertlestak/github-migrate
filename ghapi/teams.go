package ghapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

// Team contains team data
type Team struct {
	ID              int           `json:"id"`
	NodeID          string        `json:"node_id"`
	URL             string        `json:"url"`
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	Description     string        `json:"description"`
	Privacy         string        `json:"privacy"`
	Permission      string        `json:"permission"`
	MembersURL      string        `json:"members_url"`
	Members         []*User       `json:"members"`
	RepositoriesURL string        `json:"repositories_url"`
	Parent          ParentTeam    `json:"parent"`
	MembersCount    int           `json:"members_count"`
	ReposCount      int           `json:"repos_count"`
	CreatedAt       string        `json:"created_at"`
	UpdatedAt       string        `json:"updated_at"`
	Organization    Organization  `json:"organization"`
	Repositories    []*Repository `json:"repositories"`
}

// ParentTeam contains a team's parent team data
type ParentTeam struct {
	ID              int    `json:"id"`
	NodeID          string `json:"node_id"`
	URL             string `json:"url"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	Description     string `json:"description"`
	Privacy         string `json:"privacy"`
	Permission      string `json:"permission"`
	MembersURL      string `json:"members_url"`
	RepositoriesURL string `json:"repositories_url"`
}

// AllTeams lists all members in org
func AllTeams() ([]*Team, error) {
	var lp ListPages
	var ts []*Team
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Teams %+v\n", lp)
		tsl, llp, err := ListTeams(lp.Next)
		if err != nil {
			return ts, err
		}
		ts = append(ts, tsl...)
		if lp.Next == lp.Last && lp.Last > 0 {
			break
		}
		lp = llp
		if lp.Last == 0 {
			break
		}
	}
	return ts, nil
}

// ListTeams lists teams in an organization
func ListTeams(page int) ([]*Team, ListPages, error) {
	var tl []*Team
	var lp ListPages
	reqURL := "https://api.github.com/orgs/" + Org + "/teams"
	if page > 0 {
		reqURL += "?page=" + strconv.Itoa(page)
	}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return tl, lp, err
	}
	req.Header.Set("Accept", "application/vnd.github.hellcat-preview+json")
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return tl, lp, rerr
	}
	links := res.Header.Get("Link")
	lp, err = parseLinks(links)
	if err != nil {
		return tl, lp, err
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return tl, lp, rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return tl, lp, berr
	}
	jerr := json.Unmarshal(bd, &tl)
	if jerr != nil {
		return tl, lp, jerr
	}
	return tl, lp, nil
}

// GetDetails gets team details for team
func (t *Team) GetDetails() error {
	reqURL := "https://api.github.com/teams/" + strconv.Itoa(t.ID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.Printf("Get full details for team: %s\n", t.Name)
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return rerr
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return berr
	}
	jerr := json.Unmarshal(bd, &t)
	if jerr != nil {
		return jerr
	}
	return nil
}

// AllMembers lists all members in team
func (t *Team) AllMembers() ([]*User, error) {
	var lp ListPages
	var us []*User
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Members in Team %s %+v\n", t.Name, lp)
		usl, llp, err := t.ListMembers(lp.Next)
		if err != nil {
			return us, err
		}
		us = append(us, usl...)
		if lp.Next == lp.Last && lp.Last > 0 {
			break
		}
		lp = llp
		if lp.Last == 0 {
			break
		}
	}
	var lus []*User
	for _, u := range us {
		ud, uerr := u.GetDetailsLocal()
		if uerr != nil {
			return us, uerr
		}
		u = ud
		lus = append(lus, ud)
	}
	t.Members = lus
	return lus, nil
}

// ListMembers lists members in a team
func (t *Team) ListMembers(page int) ([]*User, ListPages, error) {
	var ul []*User
	var lp ListPages
	reqURL := "https://api.github.com/teams/" + strconv.Itoa(t.ID) + "/members"
	if page > 0 {
		reqURL += "?page=" + strconv.Itoa(page)
	}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return ul, lp, err
	}
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return ul, lp, rerr
	}
	links := res.Header.Get("Link")
	lp, err = parseLinks(links)
	if err != nil {
		return ul, lp, err
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return ul, lp, rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return ul, lp, berr
	}
	jerr := json.Unmarshal(bd, &ul)
	if jerr != nil {
		return ul, lp, jerr
	}
	return ul, lp, nil
}

// SaveTeamList saves a member list to a JSON file
func SaveTeamList(ts []*Team) error {
	teamListFile := path.Join(DataDir, "teams.json")
	os.Remove(teamListFile)
	log.SetOutput(os.Stdout)
	log.Printf("Saving team list to: %s\n", teamListFile)
	jd, jerr := json.Marshal(ts)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(teamListFile, jd, 0755)
}

// InviteMemberToTeam invites user to org
func (m *Membership) InviteMemberToTeam(t *Team) error {
	reqURL := "https://api.github.com/teams/" + strconv.Itoa(t.ID) + "/memberships/" + m.User.Login
	type params struct {
		Role string `json:"role"`
	}
	p := &params{
		Role: m.Role,
	}
	jd, jerr := json.Marshal(&p)
	if jerr != nil {
		return jerr
	}
	req, err := http.NewRequest("PUT", reqURL, bytes.NewBuffer(jd))
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.Printf("Invite user %s to team: %s\n", m.User.Login, t.Name)
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return rerr
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return rlerr
	}
	return nil
}

// TeamIDs returns team IDs for a membership
func (m *Membership) TeamIDs() ([]int, error) {
	var ids []int
	teamListFile := path.Join(DataDir, "teams.json")
	if _, cerr := os.Stat(teamListFile); os.IsNotExist(cerr) {
		log.Println(teamListFile, "does not exist")
		return ids, cerr
	}
	fd, ferr := ioutil.ReadFile(teamListFile)
	if ferr != nil {
		return ids, ferr
	}
	var ts []*Team
	jerr := json.Unmarshal(fd, &ts)
	if jerr != nil {
		return ids, jerr
	}
	for _, t := range ts {
		for _, u := range t.Members {
			if u.ID == m.User.ID {
				ids = append(ids, t.ID)
			}
		}
	}
	return ids, nil
}

// InviteUsersToTeams invites all users defined in teams file back to team
func InviteUsersToTeams() error {
	teamListFile := path.Join(DataDir, "teams.json")
	if _, cerr := os.Stat(teamListFile); os.IsNotExist(cerr) {
		log.Println(teamListFile, "does not exist")
		return cerr
	}
	fd, ferr := ioutil.ReadFile(teamListFile)
	if ferr != nil {
		return ferr
	}
	var ts []*Team
	jerr := json.Unmarshal(fd, &ts)
	if jerr != nil {
		return jerr
	}
	for _, t := range ts {
		for _, u := range t.Members {
			m, merr := u.GetLocalMembership()
			if merr != nil {
				return merr
			}
			ierr := m.InviteMemberToTeam(t)
			if ierr != nil {
				return ierr
			}
		}
	}
	return nil
}
