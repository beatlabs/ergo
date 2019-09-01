package main

import (
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/callback/github", githubHookHandler)
	http.ListenAndServe(":"+os.Getenv("WEBSERVER_PORT"), nil)
}
