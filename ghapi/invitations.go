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

type Invitation struct {
	ID                int    `json:"id"`
	Login             string `json:"login"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	CreatedAt         string `json:"created_at"`
	Inviter           *User  `json:"inviter"`
	TeamCount         int    `json:"team_count"`
	InvitationTeamURL string `json:"invitation_team_url"`
}

// GetAllInvitations lists all repos for a team
func GetAllInvitations() ([]*Invitation, error) {
	var lp ListPages
	var rs []*Invitation
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Org Invitations, %+v\n", lp)
		isl, llp, err := ListInvitations(lp.Next)
		rs = append(rs, isl...)
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

// ListInvitations lists pending invitations for org
func ListInvitations(page int) ([]*Invitation, ListPages, error) {
	var il []*Invitation
	var lp ListPages
	reqURL := "https://api.github.com/orgs/" + Org + "/invitations"
	if page > 0 {
		reqURL += "?page=" + strconv.Itoa(page)
	}
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return il, lp, err
	}
	req.Header.Set("Accept", "application/vnd.github.dazzler-preview+json")
	req.Header.Set("Authorization", "token "+Token)
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return il, lp, rerr
	}
	links := res.Header.Get("Link")
	lp, err = parseLinks(links)
	if err != nil {
		return il, lp, err
	}
	_, rlerr := ParseRateLimit(res)
	if rlerr != nil {
		return il, lp, rlerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return il, lp, berr
	}
	jerr := json.Unmarshal(bd, &il)
	if jerr != nil {
		return il, lp, jerr
	}
	return il, lp, nil
}

// SaveInvitations saves a membership list to a JSON file
func SaveInvitations(ls []*Invitation) error {
	invitationListFile := path.Join(DataDir, "invitations.json")
	os.Remove(invitationListFile)
	log.SetOutput(os.Stdout)
	log.Printf("Saving invitations list to: %s\n", invitationListFile)
	jd, jerr := json.Marshal(ls)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(invitationListFile, jd, 0755)
}
