package ghapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// OathToken generates an Oauth Token
func OathToken(tf string) (OathResponse, error) {
	var oar OathResponse
	or := &oathReq{
		//ClientID:     os.Getenv("OATH_CLIENT_ID"),
		ClientSecret: os.Getenv("OATH_CLIENT_SECRET"),
		Scopes:       strings.Split(os.Getenv("OATH_SCOPES"), ","),
		Fingerprint:  os.Getenv("OAUTH_FINGERPRINT"),
		Note:         os.Getenv("OAUTH_NOTE"),
	}
	jb, jerr := json.Marshal(&or)
	if jerr != nil {
		return oar, jerr
	}
	ourl := "https://api.github.com/authorizations/clients/" + os.Getenv("OATH_CLIENT_ID")
	req, err := http.NewRequest("PUT", ourl, bytes.NewReader(jb))
	if err != nil {
		return oar, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-OTP", tf)
	req.SetBasicAuth(os.Getenv("GITHUB_USERNAME"), os.Getenv("GITHUB_PASSWORD"))
	c := &http.Client{}
	res, rerr := c.Do(req)
	if rerr != nil {
		return oar, rerr
	}
	defer res.Body.Close()
	bd, berr := ioutil.ReadAll(res.Body)
	if berr != nil {
		return oar, berr
	}
	if res.StatusCode != 200 {
		return oar, errors.New(string(bd))
	}
	fmt.Println(string(bd))
	oerr := json.Unmarshal(bd, &oar)
	if oerr != nil {
		return oar, oerr
	}
	return oar, nil
}
