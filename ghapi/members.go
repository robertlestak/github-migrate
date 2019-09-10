package ghapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

// User contains GitHub user data
type User struct {
	Login            string `json:"login"`
	ID               int    `json:"id"`
	NodeID           string `json:"node_id"`
	AvatarURL        string `json:"avatar_url"`
	URL              string `json:"url"`
	HTMLURL          string `json:"html_url"`
	OrganizationsURL string `json:"organizations_url"`
	Type             string `json:"type"`
	SiteAdmin        bool   `json:"site_admin"`
	Name             string `json:"name"`
	Company          string `json:"company"`
	Blog             string `json:"blog"`
	Location         string `json:"location"`
	Email            string `json:"email"`
	Hireable         bool   `json:"hireable"`
	Bio              string `json:"bio"`
}

// AllMembers lists all members in org
func AllMembers() ([]*User, error) {
	var lp ListPages
	var us []*User
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Members %+v\n", lp)
		usl, llp, err := ListMembers(lp.Next)
		if err != nil {
			return us, err
		}
		if lp.Next == lp.Last && lp.Last > 0 {
			us = append(us, usl...)
			break
		}
		lp = llp
		if lp.Last == 0 {
			us = append(us, usl...)
			break
		}
		us = append(us, usl...)
	}
	return us, nil
}

// ListMembers lists members in an organization
func ListMembers(page int) ([]*User, ListPages, error) {
	var ul []*User
	var lp ListPages
	reqURL := "https://api.github.com/orgs/" + Org + "/members"
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
	if res.StatusCode != 200 {
		return ul, lp, errors.New(string(bd))
	}
	jerr := json.Unmarshal(bd, &ul)
	if jerr != nil {
		return ul, lp, jerr
	}
	return ul, lp, nil
}

// SaveMemberList saves a member list to a JSON file
func SaveMemberList(ls []*User) error {
	userListFile := path.Join(DataDir, "users.json")
	os.Remove(userListFile)
	log.SetOutput(os.Stdout)
	log.Printf("Saving member list to: %s\n", userListFile)
	jd, jerr := json.Marshal(ls)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(userListFile, jd, 0755)
}

// GetDetails gets memberships for all users
func (u *User) GetDetails() error {
	reqURL := "https://api.github.com/users/" + u.Login
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.Printf("Get full details for user: %s\n", u.Login)
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
	jerr := json.Unmarshal(bd, &u)
	if jerr != nil {
		return jerr
	}
	return nil
}

// GetDetailsLocal gets the user details from the local data file
func (u *User) GetDetailsLocal() (*User, error) {
	userListFile := path.Join(DataDir, "users.json")
	if _, cerr := os.Stat(userListFile); os.IsNotExist(cerr) {
		log.Println(userListFile, "does not exist")
		return u, cerr
	}
	var ul []*User
	bd, rerr := ioutil.ReadFile(userListFile)
	if rerr != nil {
		return u, rerr
	}
	jerr := json.Unmarshal(bd, &ul)
	if jerr != nil {
		return u, jerr
	}
	log.SetOutput(os.Stdout)
	log.Printf("Get full local details for user: %s\n", u.Login)
	for _, lu := range ul {
		if lu.ID == u.ID {
			u = lu
			return lu, nil
		}
	}
	return u, nil
}

// GetLocalMembership returns membership details for a user
func (u *User) GetLocalMembership() (*Membership, error) {
	var m *Membership
	memberListFile := path.Join(DataDir, "memberships.json")
	if _, cerr := os.Stat(memberListFile); os.IsNotExist(cerr) {
		log.Println(memberListFile, "does not exist")
		return m, cerr
	}
	var ms []*Membership
	bd, rerr := ioutil.ReadFile(memberListFile)
	if rerr != nil {
		return m, rerr
	}
	jerr := json.Unmarshal(bd, &ms)
	if jerr != nil {
		return m, jerr
	}
	log.SetOutput(os.Stdout)
	log.Printf("Get full membership details for user: %s\n", u.Login)
	for _, lm := range ms {
		if lm.User.Login == u.Login {
			m = lm
			return m, nil
		}
	}
	return m, nil
}

// Remove removes member from org
func (m *Membership) Remove() error {
	reqURL := "https://api.github.com/orgs/" + Org + "/members/" + m.User.Login
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.Printf("Delete user from org %s: %s\n", Org, m.User.Login)
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

// Invite invites user to org
func (m *Membership) Invite() error {
	reqURL := "https://api.github.com/orgs/" + Org + "/invitations"
	type params struct {
		InviteeID int    `json:"invitee_id,omitempty"`
		Email     string `json:"email,omitempty"`
		TeamIDs   []int  `json:"team_ids,omitempty"`
		Role      string `json:"role"`
	}
	if m.Role == "member" {
		m.Role = "direct_member"
	}
	p := &params{
		Role: m.Role,
	}
	if m.User.ID == 0 && m.User.Email != "" {
		p.Email = m.User.Email
	} else {
		p.InviteeID = m.User.ID
	}
	var terr error
	p.TeamIDs, terr = m.TeamIDs()
	if terr != nil {
		return terr
	}
	jd, jerr := json.Marshal(&p)
	if jerr != nil {
		return jerr
	}
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jd))
	if err != nil {
		return err
	}
	log.SetOutput(os.Stdout)
	log.Printf("Invite user to org: %s\n", m.User.Login)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+Token)
	req.Header.Set("Accept", "application/vnd.github.dazzler-preview+json")
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
	if res.StatusCode > 202 {
		return errors.New(string(bd))
	}
	return nil
}
