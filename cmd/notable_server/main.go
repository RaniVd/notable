package main

import (
	"encoding/json"
	"fmt"
	notable "github.com/harvesthq/notable"
	"log"
	"net/http"
	"net/url"
	"os"
)

const webHookToken = "13XIbBjtLeimhPIY36DWZvdR"
const slashCommandToken = "jINvK9gvlwQafaCR3yWlksRW"

type OKResponse struct {
	Text string `json:"text"`
}

type SummaryResponse struct {
	Notes []notable.Note `json:"notes"`
}

func getAndSetHandler(responseWriter http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	token := request.Form.Get("token")

	if validToken(token) {
		var err error

		if request.Method == "POST" {
			recordNote(request.Form)
			if viaSlashCommand(token) {
				responseWriter.Write([]byte("Got it!"))
			} else {
				respondWithJSON(responseWriter, OKResponse{"Got it!"})
			}
		} else {
			respondWithJSON(responseWriter, SummaryResponse{notable.Notes()})
		}

		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Printf("Invalid token received: %s\n", token)
		http.Error(responseWriter, "Invalid token", http.StatusForbidden)
		return
	}
}

func clearHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		notable.Reset()
	}
}

func emailHandler(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "text/html")
	responseWriter.Write([]byte(notable.Email()))
}

func recordNote(form url.Values) {
	user_id := form.Get("user_id")
	category := form.Get("trigger_word")
	text := form.Get("text")
	channel := form.Get("channel_name")
	slackToken := os.Getenv("SLACK_API_TOKEN")

	notable.Record(user_id, category, text, channel, slackToken)
}

func validToken(token string) bool {
	return token == webHookToken || token == slashCommandToken
}

func viaSlashCommand(token string) bool {
	return token == slashCommandToken
}

func respondWithJSON(responseWriter http.ResponseWriter, response interface{}) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.Write(responseJSON)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/email", emailHandler)
	http.HandleFunc("/clear", clearHandler)
	http.HandleFunc("/", getAndSetHandler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}