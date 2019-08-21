package ghapi

import (
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Org is the GitHub organization
	Org string
	// DataDir is the directory in which data is stored
	DataDir string
	// Token is the GitHub auth token. Must have proper access to org.
	Token string
)

type oathReq struct {
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Fingerprint  string   `json:"fingerprint,omitempty"`
	Note         string   `json:"note,omitempty"`
}

// OathResponse returns the response for an Oauth Request
type OathResponse struct {
	ID             int      `json:"id"`
	URL            string   `json:"url"`
	App            oauthApp `json:"app"`
	Token          string   `json:"token"`
	HashedToken    string   `json:"hashed_token"`
	TokenLastEight string   `json:"token_last_eight"`
	Note           string   `json:"note"`
	NoteURL        string   `json:"note_url"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	Scopes         []string `json:"scopes"`
	Fingerprint    string   `json:"fingerprinte"`
}

type oauthApp struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	ClientID string `json:"client_id"`
}

// GitHubError handles an error returned by GitHub
type GitHubError struct {
	Message          string                `json:"message"`
	Errors           []GitHubResponseError `json:"errors"`
	DocumentationURL string                `json:"documentation_url"`
}

// GitHubResponseError handles a list of GH errors
type GitHubResponseError struct {
	Resource string `json:"resource"`
	Code     string `json:"code"`
	Field    string `json:"field"`
}

// RateLimit contains the rate limit data
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int
}

// ListPages returns the pages in the list query
type ListPages struct {
	Prev int
	Next int
	Last int
}

// ParseRateLimit parses the rate limit from headers
func ParseRateLimit(res *http.Response) (*RateLimit, error) {
	var rl *RateLimit
	var err error
	var rlim int
	var rrem int
	var rres int
	rlim, err = strconv.Atoi(res.Header.Get("X-RateLimit-Limit"))
	if err != nil {
		return rl, err
	}
	rrem, err = strconv.Atoi(res.Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		return rl, err
	}
	rres, err = strconv.Atoi(res.Header.Get("X-RateLimit-Reset"))
	if err != nil {
		return rl, err
	}
	rl = &RateLimit{
		Limit:     rlim,
		Remaining: rrem,
		Reset:     rres,
	}
	if rl.Remaining <= 50 {
		log.Printf("Rate Limit Reached, sleeping for %d ms", rl.Reset)
		time.Sleep(time.Millisecond * time.Duration(rl.Reset))
	}
	return rl, nil
}

func pageFromLink(link string) (int, error) {
	nr := regexp.MustCompile("/?page=.*>")
	np := nr.FindString(link)
	nextPageStr := strings.Replace(np, "page=", "", -1)
	nextPageStr = strings.Replace(nextPageStr, ">", "", -1)
	intPage, nerr := strconv.Atoi(nextPageStr)
	if nerr != nil {
		return intPage, nerr
	}
	return intPage, nil
}

func parseLinks(links string) (ListPages, error) {
	var lp ListPages
	var err error
	lar := strings.Split(links, ",")
	for _, v := range lar {
		if strings.Contains(v, "rel=\"next\"") {
			var np int
			np, err = pageFromLink(v)
			lp.Next = np
		} else if strings.Contains(v, "rel=\"prev\"") {
			var pp int
			pp, err = pageFromLink(v)
			lp.Prev = pp
		} else if strings.Contains(v, "rel=\"last\"") {
			var lap int
			lap, err = pageFromLink(v)
			lp.Last = lap
		}
	}
	return lp, err
}
