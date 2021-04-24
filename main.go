package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"encoding/json"
)

var (
	port = "8080"
)

func cowHandler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("SLACK_SLASH_COMMAND_VERIFICATION_TOKEN")
	if "" == token {
		panic("SLACK_SLASH_COMMAND_VERIFICATION_TOKEN is not set!")
	}

	log.Printf("in handler")
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if token != r.FormValue("token") {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	serviceName := strings.Replace(r.FormValue("text"), "\r", "", -1)
	statusResponse := status(serviceName)

	jsonResp, _ := json.Marshal(struct {
		Type string `json:"response_type"`
		Text string `json:"text"`
	}{
		Type: "in_channel",
		Text: fmt.Sprintf("```%s```", statusResponse),
	})

	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResp))
}

func main() {
	http.HandleFunc("/", cowHandler)
	log.Fatalln(http.ListenAndServe(":"+port, nil))
}
