package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type RocketMessage struct {
	UserID string      `json:"user_id"`
	Text   string      `json:"text"`
	Bot    interface{} `json:"bot,omitempty"`
}

type RasaMessage struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

type ResponseToRocket struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
	Text      string `json:"text"`
	ImageURL  string `json:"image_url"`
	Color     string `json:"color"`
}

func main() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received a new request from Rocket.Chat")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		log.Printf("Rocket received body: %s", string(body))

		var msg RocketMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Error parsing JSON request body: %v", err)
			http.Error(w, "Error parsing JSON request body", http.StatusBadRequest)
			return
		}

		if bot, ok := msg.Bot.(bool); ok && !bot {
			log.Println("Message is not from a bot")
		} else if botMap, ok := msg.Bot.(map[string]interface{}); ok {
			log.Printf("Message is from bot with id: %v", botMap["i"])
			w.WriteHeader(http.StatusNoContent)
			return
		}

		log.Printf("Received message: UserID=%s, Text=%s", msg.UserID, msg.Text)

		rasaMsg := RasaMessage{
			Sender:  msg.UserID,
			Message: msg.Text,
		}
		rasaMsgJSON, err := json.Marshal(rasaMsg)
		if err != nil {
			log.Printf("Error generating JSON for Rasa: %v", err)
			http.Error(w, "Error generating JSON", http.StatusInternalServerError)
			return
		}

		resp, err := http.Post("http://localhost:5005/webhooks/rest/webhook", "application/json", bytes.NewBuffer(rasaMsgJSON))
		if err != nil {
			log.Printf("Error sending request to Rasa: %v", err)
			http.Error(w, "Error sending request to Rasa", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response from Rasa: %v", err)
			http.Error(w, "Error reading response from Rasa", http.StatusInternalServerError)
			return
		}

		log.Printf("Received response from Rasa: %s", string(response))
		var rasaResponse []map[string]interface{}
		err = json.Unmarshal(response, &rasaResponse)
		if err != nil {
			return
		}

		var responseText string
		if len(rasaResponse) > 0 {
			responseText = rasaResponse[0]["text"].(string)
		}

		rocketResponse := ResponseToRocket{
			Text:        responseText,
			Attachments: []Attachment{},
		}
		rocketResponseJSON, err := json.Marshal(rocketResponse)
		if err != nil {
			log.Printf("Error generating response JSON for Rocket.Chat: %v", err)
			http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(rocketResponseJSON)
		if err != nil {
			return
		}
	})

	log.Println("Server is running on port 5002")
	log.Fatal(http.ListenAndServe(":5002", nil))
}
