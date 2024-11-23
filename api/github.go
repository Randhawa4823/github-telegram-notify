package api

import (
	"github-webhook/GithubEvent/str"
	"net/http"
)

// GitHub processes GitHub webhooks
func GitHub(w http.ResponseWriter, r *http.Request) {
	str.GitHubWebhook(w, r)
}
