package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"
)

const (
	telegramAPIBaseURL = "https://api.telegram.org/bot"
	maxRetries         = 5
)

var (
	botToken     string
	githubToken  string
	shutdown     chan os.Signal
	httpClient   *http.Client
	githubApiUrl string
)

type Update struct {
	ID          int
	Message     *struct{ Text string }
	ChannelPost *struct{ Text string }
}

func main() {
	botToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	githubToken = os.Getenv("GITHUB_TOKEN")
	githubApiUrl = os.Getenv("GITHUB_API_URL")

	if botToken == "" || githubToken == "" || githubApiUrl == "" {
		log.Fatal("Please provide required environment variables: TELEGRAM_BOT_TOKEN, API_URL, GITHUB_TOKEN, GITHUB_REPOSITORY")
	}

	httpClient = configureHTTPClient(time.Second * 10)

	shutdown = make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go pollUpdates()

	<-shutdown
	log.Println("Shutting down...")
}

func configureHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

func pollUpdates() {
	offset := 0

	for {
		select {
		case <-shutdown:
			return
		default:
			updates, err := getUpdates(offset)
			if err != nil {
				log.Println("Failed to get updates:", err)
				retryPollUpdates(&offset)
			} else {
				for _, update := range updates {
					offset = update.ID + 1

					var text string
					if update.Message != nil {
						text = update.Message.Text
					} else if update.ChannelPost != nil {
						text = update.ChannelPost.Text
					}

					url, tags := parseURLAndTags(text)
					title, _ := getTitle(url)

					if url != "" {
						err := sendPOSTRequest(url, tags, title)
						if err != nil {
							log.Println("Error sending POST request:", err)
						} else {
							log.Println("POST request sent successfully!")
						}
					}
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func retryPollUpdates(offset *int) {
	for i := 0; i < maxRetries; i++ {
		sleepDuration := time.Duration(1<<i) * time.Second
		log.Printf("Retrying after %s...", sleepDuration)
		time.Sleep(sleepDuration)

		_, err := getUpdates(*offset)
		if err == nil {
			return
		}
	}

	log.Printf("Exceeded maximum retries. Exiting...")
	os.Exit(1)
}

func getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d", telegramAPIBaseURL, botToken, offset)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Result []Update `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func getTitle(url string) (string, error) {
	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	// Traverse the HTML tree and find the title tag
	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "title" {
			return n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			title := f(c)
			if title != "" {
				return title
			}
		}
		return ""
	}

	return strings.TrimSpace(f(doc)), nil
}

func parseURLAndTags(text string) (string, []string) {
	var url string
	var tags []string

	words := strings.Fields(text)

	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			// Remove the leading "#" from the tag
			tag := strings.TrimPrefix(word, "#")
			tags = append(tags, tag)
		} else if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
			url = word
			break
		}
	}

	return url, tags
}

type Dispatch struct {
	EventType     string        `json:"event_type"`
	ClientPayload ClientPayload `json:"client_payload"`
}

type ClientPayload struct {
	URL   string   `json:"url"`
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

func sendPOSTRequest(url string, tags []string, title string) error {
	data := Dispatch{
		EventType: "webhook",
		ClientPayload: ClientPayload{
			URL:   url,
			Title: title,
			Tags:  tags,
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	// Create the request
	req, err := http.NewRequest("POST", githubApiUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}

	// Add the necessary headers
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", githubToken)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	defer resp.Body.Close()

	// Print the response
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Response: %s", body)
	return nil
}
