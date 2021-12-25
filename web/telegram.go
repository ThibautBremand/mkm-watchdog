package web

import (
	"bytes"
	"fmt"
	"net/http"
)

// SendTelegramMessage sends the given message to the Telegram bot which is bound to the given token and chatID.
func SendTelegramMessage(token string, chatID string, message string) error {
	data := fmt.Sprintf("{\"chat_id\": \"%s\", \"text\": \"%s\"}", chatID, message)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", token), bytes.NewBufferString(data))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make Telegram request: %v", err)
	}
	defer resp.Body.Close()

	if resp == nil {
		return fmt.Errorf("no response from Telegram request")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram responded with status code %v", resp.StatusCode)
	}

	return nil
}
