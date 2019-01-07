package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/omidnikta/logrus"
	"google.golang.org/api/googleapi/transport"
	youtube "google.golang.org/api/youtube/v3"
)

// BodyMsg message messenger
type BodyMsg struct {
	Object string      `json:"object"`
	Entry  []messaging `json:"entry"`
}

type messaging struct {
	ID        string        `json:"id"`
	Messaging []messagingEl `json:"messaging"`
	Time      int           `json:"time"`
}

type messagingEl struct {
	Message   message   `json:"message"`
	PostBack  message   `json:"postback"`
	Recipient recipient `json:"recipient"`
	Sender    sender    `json:"sender"`
	Timestamp int       `json:"timestamp"`
}

type message struct {
	Mid  string `json:"mid"`
	Seq  int    `json:"seq"`
	Text string `json:"text"`
}

type serverMessage struct {
	Text       string      `json:"text,omitempty"`
	Attachment *attachment `json:"attachment,omitempty"`
}

type recipient struct {
	ID string `json:"id"`
}

type sender struct {
	ID string `json:"id"`
}

type response struct {
	Recipient recipient     `json:"recipient"`
	Message   serverMessage `json:"message"`
}

type privacyData struct {
	Domain   string
	Business string
	City     string
	Country  string
}

type attachment struct {
	AttachmentType string  `json:"type"`
	Payload        payload `json:"payload"`
}

type payload struct {
	TemplateType string    `json:"template_type"`
	Elements     []element `json:"elements"`
}

type element struct {
	Title         string        `json:"title"`
	ImageURL      string        `json:"image_url"`
	DefaultAction defaultAction `json:"default_action"`
	Buttons       []button      `json:"buttons"`
}

type defaultAction struct {
	DefaultActionType   string `json:"type"`
	URL                 string `json:"url"`
	WebViewHeightRatio  string `json:"webview_height_ratio"`
	MessengerExtensions bool   `json:"messenger_extensions"`
}

type button struct {
	ButtonType          string `json:"type"`
	URL                 string `json:"url"`
	Title               string `json:"title"`
	MessengerExtensions bool   `json:"messenger_extensions"`
	WebViewHeightRatio  string `json:"webview_height_ratio"`
}

func hello(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Greetings from mi AI %s!", r.URL.Path[1:])
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

			handleMessage(webhookEvent.Sender.ID, webhookEvent.Message)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", "EVENT_RECEIVED")
	} else {
		http.Error(w, "Message error", http.StatusBadRequest)
	}
}

func privacyHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	privacyData := privacyData{
		os.Getenv("DOMAIN"),
		os.Getenv("BUSINESS"),
		os.Getenv("CITY"),
		os.Getenv("COUNTRY"),
	}

	renderTemplate(w, "template/privacy_policy", privacyData)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data privacyData) {
	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, data)
}

// Handles messages events
func handleMessage(senderPsid string, receivedMessage message) {
	var res response

	if receivedMessage.Text != "" {
		searchResults := youtubeSearchAPI(receivedMessage.Text)

		if len(searchResults) > 0 {

			var (
				elements []element
				buttons  []button
			)

			for _, item := range searchResults {
				elements = append(elements, element{
					Title:    item.Snippet.Title,
					ImageURL: item.Snippet.Thumbnails.Default.Url,
					DefaultAction: defaultAction{
						DefaultActionType:   "web_url",
						URL:                 "https://www.youtube.com/watch?v=" + item.Id.VideoId,
						WebViewHeightRatio:  "tall",
						MessengerExtensions: true,
					},
					Buttons: append(buttons, button{
						ButtonType:          "web_url",
						URL:                 "https://www.youtube.com/watch?v=" + item.Id.VideoId,
						Title:               "Download",
						MessengerExtensions: true,
						WebViewHeightRatio:  "tall",
					}),
				})
			}

			res = response{
				Recipient: recipient{
					ID: senderPsid,
				},
				Message: serverMessage{
					Attachment: &attachment{
						AttachmentType: "template",
						Payload: payload{
							TemplateType: "generic",
							Elements:     elements,
						},
					},
				},
			}

		} else {
			res = response{
				Recipient: recipient{
					ID: senderPsid,
				},
				Message: serverMessage{
					Text: "Aucun résultat n'a été trouvé pour " + receivedMessage.Text,
				},
			}
		}

		callSendAPI(senderPsid, res)
	}
}

// Handles messaging_postbacks events
func handlePostback(senderPsid string, receivedPostback message) {

}

// Sends response messages via the Send API
func callSendAPI(senderPsid string, res response) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(res)

	if err != nil {
		logrus.Errorf("Failed to encode JSON: %v.", err)
		return
	}

	url := "https://graph.facebook.com/v2.6/me/messages"
	req, err := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", "Bearer "+os.Getenv("PAGE_ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Body:", string(respBody))
}

func youtubeSearchAPI(video string) []*youtube.SearchResult {
	client := &http.Client{
		Transport: &transport.APIKey{Key: os.Getenv("YOUTUBE_DATA_API_KEY")},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(video).
		MaxResults(5)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	return response.Items
}

func printIDs(sectionName string, matches map[string]string) {
	fmt.Printf("%v:\n", sectionName)
	for id, title := range matches {
		fmt.Printf("[%v] %v\n", id, title)
	}
	fmt.Printf("\n\n")
}

// Sends response messages via the Send API
func facebookSendAPI(senderPsid string, res response) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(res)

	if err != nil {
		logrus.Errorf("Failed to encode JSON: %v.", err)
		return
	}

	url := "https://graph.facebook.com/v2.6/me/messages"
	req, err := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", "Bearer "+os.Getenv("PAGE_ACCESS_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Body:", string(respBody))
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
	router.GET("/privacy", privacyHandler)
	router.GET("/webhook", verifyHook)
	router.POST("/webhook", apiHook)

	error := http.ListenAndServeTLS(":443", certPath, certKeyPath, router)
	if error != nil {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
