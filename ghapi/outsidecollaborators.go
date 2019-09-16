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

// GetAllOutsideCollaborators lists all repos for a team
func GetAllOutsideCollaborators() ([]*User, error) {
	var lp ListPages
	var us []*User
	for lp.Next <= lp.Last {
		if lp.Next == 0 {
			lp.Next = 1
		}
		log.SetOutput(os.Stdout)
		log.Printf("Listing Org Outside Collaborators, %+v\n", lp)
		usl, llp, err := ListOutsideCollaborators(lp.Next)
		for _, u := range usl {
			gerr := u.GetDetails()
			if gerr != nil {
				return us, gerr
			}
		}
		us = append(us, usl...)
		if err != nil {
			return us, err
		}
		if lp.Next == lp.Last && lp.Last > 0 {
			break
		}
		lp = llp
		if lp.Last == 0 {
			break
		}
	}
	return us, nil
}

// ListOutsideCollaborators lists outside collaborators for org
func ListOutsideCollaborators(page int) ([]*User, ListPages, error) {
	var ul []*User
	var lp ListPages
	reqURL := "https://api.github.com/orgs/" + Org + "/outside_collaborators"
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

// SaveOutsideCollaborators saves an outside collaborator list to a JSON file
func SaveOutsideCollaborators(ls []*User) error {
	collaboratorList := path.Join(DataDir, "outside_collaborators.json")
	os.Remove(collaboratorList)
	log.SetOutput(os.Stdout)
	log.Printf("Saving outside collaborator list to: %s\n", collaboratorList)
	jd, jerr := json.Marshal(ls)
	if jerr != nil {
		return jerr
	}
	return ioutil.WriteFile(collaboratorList, jd, 0755)
}
