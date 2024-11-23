package api

import (
	"github-webhook/GithubEvent/str"
	"net/http"
)

// Handler processes HTTP requests to the / endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	str.Home(w, r)
}
