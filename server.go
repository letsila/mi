package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Greetings from mimi.ai %s!", r.URL.Path[1:])
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	certPath := os.Getenv("CERT_PATH")
	certKeyPath := os.Getenv("CERT_KEY_PATH")

	http.HandleFunc("/", handler)

	error := http.ListenAndServeTLS(":443", certPath, certKeyPath, nil)
	if error != nil {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
