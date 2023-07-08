package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type MockBot struct {
	ChatID  int64
	Message string
}

func (mb *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	msg := c.(tgbotapi.MessageConfig) // cast to the concrete type
	mb.ChatID = msg.ChatID
	mb.Message = msg.Text
	return tgbotapi.Message{}, nil
}

func TestGetTitleFromURL(t *testing.T) {
	title, err := getTitleFromURL("https://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if title != "Example Domain" { // Assuming example.com's title is "Example Domain"
		t.Errorf("expected 'Example Domain', got %v", title)
	}
}

func TestConcatenateHashtags(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "single hashtag",
			input:    []string{"#test"},
			expected: "test",
		},
		{
			name:     "multiple hashtags",
			input:    []string{"#test1", "#test2"},
			expected: "test1,test2",
		},
		{
			name:     "no hashtags",
			input:    []string{},
			expected: "unsorted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := concatenateHashtags(tc.input)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestTriggerWorkflowRun(t *testing.T) {
	client := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			}, nil
		},
	}

	payload := URLInfo{
		URL:      "https://www.example.com",
		Title:    "Example Title",
		Hashtags: "tag1,tag2,tag3",
	}

	err := triggerWorkflowRun(client, "fakeToken", "https://api.github.com/repos/owner/repo/dispatches", payload)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
func TestSendMessage(t *testing.T) {
	mockBot := &MockBot{}
	chatID := int64(123456)
	message := "Test message"

	sendMessage(mockBot, chatID, message)

	if mockBot.ChatID != chatID || mockBot.Message != message {
		t.Errorf("Expected ChatID %v and Message %v, but got %v and %v", chatID, message, mockBot.ChatID, mockBot.Message)
	}
}
