package main

import (
	"github-webhook/GithubEvent/config"
	"github-webhook/GithubEvent/str"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", str.Home)
	http.HandleFunc("/github", str.GitHubWebhook)

	port := config.Port
	if port == "" {
		port = "3000"
	}
	
	log.Printf("Server running on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
