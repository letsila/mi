package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/omidnikta/logrus"
)

// {"object": "page", "entry": [{"messaging": [{"message": "TEST_MESSAGE"}]}]}

// BodyMsg message messenger
type BodyMsg struct {
	Object string      `json:"object"`
	Entry  []messaging `json:"entry"`
}

type messaging struct {
	Messaging []message `json:"messaging"`
}

type message struct {
	Message string `json:"message"`
}

func hello(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Greetings from mimi ai %s!", r.URL.Path[1:])
}

func verifyHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	verifyToken := os.Getenv("VERIFY_TOKEN")

	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode != "" && token != "" {
		if mode == "subscribe" && token == verifyToken {
			fmt.Println("WEBHOOK_VERIFIED")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", challenge)
		} else {
			http.Error(w, "Challenge failure", http.StatusBadRequest)
		}
	}
}

func apiHook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body := BodyMsg{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		logrus.Errorf("Failed to decode JSON: %v.", err)
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if body.Object == "page" {
		for _, entry := range body.Entry {
			webhookEvent := entry.Messaging[0]

			fmt.Println(webhookEvent)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "EVENT_RECEIVED")
	} else {
		http.Error(w, "Message error", http.StatusBadRequest)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	certPath := os.Getenv("CERT_PATH")
	certKeyPath := os.Getenv("CERT_KEY_PATH")

	router := httprouter.New()
	router.GET("/", hello)
	router.GET("/webhook", verifyHook)
	router.POST("/webhook", apiHook)

	error := http.ListenAndServeTLS(":443", certPath, certKeyPath, router)
	if error != nil {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
