package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"

	"github.com/umg/devops-github-migrate/ghapi"
)

func pullRepositories() {
	rs, merr := ghapi.OrgRepositories()
	if merr != nil {
		log.Fatal(merr)
	}
	ghapi.SaveRepositories(rs)
}

func pullMembership() {
	ms, merr := ghapi.GetAllMembership()
	if merr != nil {
		log.Fatal(merr)
	}
	ghapi.SaveMembership(ms)
}

func pullInvitations() {
	is, merr := ghapi.GetAllInvitations()
	if merr != nil {
		log.Fatal(merr)
	}
	ghapi.SaveInvitations(is)
}

func pullOutsideCollaborators() {
	cs, merr := ghapi.GetAllOutsideCollaborators()
	if merr != nil {
		log.Fatal(merr)
	}
	ghapi.SaveOutsideCollaborators(cs)
}

func pullUsers() {
	us, uerr := ghapi.AllMembers()
	if uerr != nil {
		log.Fatal(uerr)
	}
	for _, u := range us {
		derr := u.GetDetails()
		if derr != nil {
			log.Fatal(derr)
		}
	}
	ghapi.SaveMemberList(us)
}

func pullTeams() {
	ts, terr := ghapi.AllTeams()
	if terr != nil {
		log.Fatal(terr)
	}
	for _, t := range ts {
		derr := t.GetDetails()
		if derr != nil {
			log.Fatal(derr)
		}
		trs, terr := t.TeamRepositories()
		if terr != nil {
			log.Fatal(terr)
		}
		tms, merr := t.AllMembers()
		if merr != nil {
			log.Fatal(merr)
		}
		t.Repositories = trs
		t.Members = tms
	}
	ghapi.SaveTeamList(ts)
}

func migrateUser(u ghapi.User) error {
	m, err := u.GetLocalMembership()
	if err != nil {
		return err
	}
	if m.URL != "" {
		rerr := m.Remove()
		if rerr != nil {
			return rerr
		}
	} else {
		err := u.GetDetails()
		if err != nil {
			return err
		}
		m.User = u
	}
	ierr := m.Invite()
	if ierr != nil {
		return ierr
	}
	return nil
}

func removeUser(u ghapi.User) error {
	m, err := u.GetLocalMembership()
	if err != nil {
		return err
	}
	if m.URL != "" {
		rerr := m.Remove()
		if rerr != nil {
			return rerr
		}
	} else {
		err := u.GetDetails()
		if err != nil {
			return err
		}
		m.User = u
	}
	return nil
}

func printUserList(ul []*ghapi.User, d string) {
	for _, lu := range ul {
		val := reflect.ValueOf(lu).Elem()
		for i := 0; i < val.NumField(); i++ {
			valueField := val.Field(i)
			typeField := val.Type().Field(i)
			tag := typeField.Tag.Get("json")
			if tag == d && valueField.String() != "" {
				fmt.Println(valueField)
			}
		}
	}
}

func printUsers(d string) error {
	userListFile := path.Join(*dataDir, "users.json")
	if _, cerr := os.Stat(userListFile); os.IsNotExist(cerr) {
		log.Println(userListFile, "does not exist")
		return cerr
	}
	var ul []*ghapi.User
	bd, rerr := ioutil.ReadFile(userListFile)
	if rerr != nil {
		return rerr
	}
	jerr := json.Unmarshal(bd, &ul)
	if jerr != nil {
		return jerr
	}
	printUserList(ul, d)
	return nil
}

func printUsersInTeam(d string, t string) error {
	teamListFile := path.Join(*dataDir, "teams.json")
	if _, cerr := os.Stat(teamListFile); os.IsNotExist(cerr) {
		log.Println(teamListFile, "does not exist")
		return cerr
	}
	var tl []*ghapi.Team
	bd, rerr := ioutil.ReadFile(teamListFile)
	if rerr != nil {
		return rerr
	}
	jerr := json.Unmarshal(bd, &tl)
	if jerr != nil {
		return jerr
	}
	var ul []*ghapi.User
	for _, lt := range tl {
		if lt.Slug == t {
			for _, lu := range lt.Members {
				ul = append(ul, lu)
			}
		}
	}
	printUserList(ul, d)
	return nil
}

func printTeams() error {
	teamListFile := path.Join(*dataDir, "teams.json")
	if _, cerr := os.Stat(teamListFile); os.IsNotExist(cerr) {
		log.Println(teamListFile, "does not exist")
		return cerr
	}
	var tl []*ghapi.Team
	bd, rerr := ioutil.ReadFile(teamListFile)
	if rerr != nil {
		return rerr
	}
	jerr := json.Unmarshal(bd, &tl)
	if jerr != nil {
		return jerr
	}
	for _, lt := range tl {
		fmt.Println(lt.Slug)
	}
	return nil
}

func checkAndPull() {
	var pullReq bool
	if _, err := os.Stat(path.Join(ghapi.DataDir, "memberships.json")); os.IsNotExist(err) {
		pullReq = true
	}
	if _, err := os.Stat(path.Join(ghapi.DataDir, "teams.json")); os.IsNotExist(err) {
		pullReq = true
	}
	if _, err := os.Stat(path.Join(ghapi.DataDir, "users.json")); os.IsNotExist(err) {
		pullReq = true
	}
	if pullReq {
		pullAll()
	}
}
