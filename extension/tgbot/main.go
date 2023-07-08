package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	urlPattern     = regexp.MustCompile(`(http|https)://[^\s]+`)
	hashtagPattern = regexp.MustCompile(`#[^\s]+`)
)

// Struct to hold URL information
type URLInfo struct {
	URL      string
	Title    string
	Hashtags string
}

func main() {
	// Retrieve the Telegram Bot Token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	// Retrieve the API Token from environment variable
	apiToken := os.Getenv("GITHUB_TOKEN")
	if apiToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	// Retrieve the Workflow URL from environment variable
	workflowURL := os.Getenv("GITHUB_API_URL")
	if workflowURL == "" {
		log.Fatal("GITHUB_API_URL environment variable not set")
	}

	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create new Bot API: %v", err)
	}

	// Enable debug mode
	bot.Debug = true

	// Create a new update configuration
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Get updates from the bot
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Process incoming updates
	for update := range updates {
		if update.Message == nil {
			log.Println("Received nil Message from the update")
			continue
		}

		// Get the message text
		messageText := update.Message.Text

		// Check if the message is a forward from another chat, bot, or channel
		if update.Message.ForwardFrom != nil || update.Message.ForwardFromChat != nil {
			messageText = fmt.Sprintf("Forwarded message: %s", messageText)
		}

		// Find URLs in the message
		urls := urlPattern.FindAllString(messageText, -1)

		// Find hashtags in the message
		hashtags := concatenateHashtags(hashtagPattern.FindAllString(messageText, -1))

		// Process each URL
		for _, url := range urls {
			// Grab the title of the URL
			title, err := getTitleFromURL(url)
			if err != nil {
				log.Printf("Error getting title for URL %s: %v\n", url, err)
				continue
			}

			// Create URLInfo struct
			urlInfo := URLInfo{
				URL:      url,
				Title:    title,
				Hashtags: hashtags,
			}

			// Log URLInfo
			fmt.Printf("URL: %s, Title: %s, Hashtags: %v\n", urlInfo.URL, urlInfo.Title, urlInfo.Hashtags)

			// Trigger GitHub Actions workflow
			err = triggerWorkflowRun(apiToken, workflowURL, urlInfo)
			if err != nil {
				log.Printf("Error triggering workflow: %v\n", err)
				// Send a message to the user indicating the error
				sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("Failed to send URL '%s' to GitHub", urlInfo.URL))
			} else {
				// Send a message to the user indicating success
				sendMessage(bot, update.Message.Chat.ID, fmt.Sprintf("URL '%s' sent to GitHub successfully", urlInfo.URL))
			}
		}
	}
}

// getTitleFromURL retrieves the title of a webpage given its URL
func getTitleFromURL(url string) (string, error) {
	// Make an HTTP GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Use goquery to parse the HTML response
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", err
	}

	// Extract the title tag value
	title := doc.Find("title").First().Text()

	// Clean up the title by removing leading/trailing whitespace and newlines
	title = strings.TrimSpace(title)

	return title, nil
}

// triggerWorkflowRun triggers a GitHub Actions workflow run with the provided payload
func triggerWorkflowRun(apiToken string, workflowURL string, payload URLInfo) error {
	// Create the request payload
	type RequestPayload struct {
		EventType     string `json:"event_type"`
		ClientPayload struct {
			URL   string `json:"url"`
			Title string `json:"title"`
			Tags  string `json:"tags"`
		} `json:"client_payload"`
	}

	requestPayload := RequestPayload{
		EventType: "webhook",
		ClientPayload: struct {
			URL   string "json:\"url\""
			Title string "json:\"title\""
			Tags  string "json:\"tags\""
		}{},
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return err
	}

	// Create the HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest(http.MethodPost, workflowURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	// Set the necessary headers for authentication and content type
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	// Log the request information
	log.Printf("GitHub Request URL: %s", request.URL.String())
	log.Printf("GitHub Request Method: %s", request.Method)
	log.Printf("GitHub Request Body: %s", jsonPayload)

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// Log the response information
	log.Printf("GitHub Response Status: %s", response.Status)
	log.Printf("GitHub Response Headers: %v", response.Header)
	// Read the response body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.Printf("GitHub Response Body: %s", responseBody)

	// Check the response status code
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("workflow run failed with status code: %d", response.StatusCode)
	}

	return nil
}

// sendMessage sends a message to the specified chat ID using the Telegram bot
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func concatenateHashtags(hashtags []string) string {
	if len(hashtags) > 0 {
		return strings.Join(hashtags, ",")
	}
	return "unsorted"
}
