package external

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type TelegramClient interface {
	SendMessage(chatId int64, text string) (TelegramResponse, error)
}

type TelegramClientImpl struct {
	Endpoint string
	Botname  string
}

func (t *TelegramClientImpl) SendMessage(chatId int64, text string) (TelegramResponse, error) {
	sanitized := url.QueryEscape(text)
	reqURL := fmt.Sprintf("%s/sendmessage?chat_id=%d&text=%s&parse_mode=markdown&disable_web_page_preview=true&disable_notification=true", t.Endpoint, chatId, sanitized)

	resp, err := http.Get(reqURL)
	if err != nil {
		log.Println("Failed to send telegram message", err)
		return TelegramResponse{}, err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	err = json.NewDecoder(resp.Body).Decode(&telegramResp)
	if err != nil {
		log.Println("Failed to decode telegram send message response", err)
		return TelegramResponse{}, err
	}
	return telegramResp, nil
}

type TelegramResponse struct {
	Ok     bool           `json:"ok"`
	Result TelegramResult `json:"result"`
}

type TelegramResult struct {
	MessageID int           `json:"message_id"`
	Chat      TelegramChat  `json:"chat"`
	From      *TelegramUser `json:"from"`
	Date      int           `json:"date"`
	Text      *string       `json:"text"`
	Caption   *string       `json:"caption"`
}

type TelegramUser struct {
	ID        int64   `json:"id"`
	IsBot     bool    `json:"is_bot"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
}

type TelegramChat struct {
	ID        int64   `json:"id"`
	Type      string  `json:"type"`
	Title     *string `json:"title"`
	FirstName *string `json:"first_name"`
}
