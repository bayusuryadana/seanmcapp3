package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// uploadTimeout is generous because a multipart video upload can push tens of MB.
const uploadTimeout = 120 * time.Second

type TelegramClient interface {
	SendMessage(chatId int64, text string) (TelegramResponse, error)
	SendPhoto(chatId int64, photoURL, caption string) (TelegramResponse, error)
	SendVideo(chatId int64, videoURL, caption string) (TelegramResponse, error)
	SendVideoUpload(chatId int64, data []byte, filename, caption string) (TelegramResponse, error)
}

type TelegramClientImpl struct {
	Endpoint     string
	Botname      string
	client       *http.Client
	uploadClient *http.Client
}

func NewTelegramClient(endpoint, botname string) *TelegramClientImpl {
	return &TelegramClientImpl{
		Endpoint:     endpoint,
		Botname:      botname,
		client:       newHTTPClient(),
		uploadClient: &http.Client{Timeout: uploadTimeout},
	}
}

func (t *TelegramClientImpl) SendMessage(chatId int64, text string) (TelegramResponse, error) {
	sanitized := url.QueryEscape(text)
	reqURL := fmt.Sprintf("%s/sendmessage?chat_id=%d&text=%s&parse_mode=markdown&disable_web_page_preview=true&disable_notification=true", t.Endpoint, chatId, sanitized)

	resp, err := t.client.Get(reqURL)
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

func (t *TelegramClientImpl) SendPhoto(chatId int64, photoURL, caption string) (TelegramResponse, error) {
	sanitized := url.QueryEscape(caption)
	reqURL := fmt.Sprintf("%s/sendphoto?chat_id=%d&photo=%s&caption=%s&parse_mode=markdown&disable_notification=true", t.Endpoint, chatId, url.QueryEscape(photoURL), sanitized)

	resp, err := t.client.Get(reqURL)
	if err != nil {
		log.Println("Failed to send telegram photo", err)
		return TelegramResponse{}, err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	err = json.NewDecoder(resp.Body).Decode(&telegramResp)
	if err != nil {
		log.Println("Failed to decode telegram send photo response", err)
		return TelegramResponse{}, err
	}
	return telegramResp, nil
}

// SendVideo asks Telegram to fetch the video from a remote URL. Telegram caps
// remote-URL videos at ~20MB; larger files come back with Ok=false and should be
// retried via SendVideoUpload.
func (t *TelegramClientImpl) SendVideo(chatId int64, videoURL, caption string) (TelegramResponse, error) {
	reqURL := fmt.Sprintf("%s/sendvideo?chat_id=%d&video=%s&caption=%s&parse_mode=markdown&disable_notification=true", t.Endpoint, chatId, url.QueryEscape(videoURL), url.QueryEscape(caption))

	resp, err := t.client.Get(reqURL)
	if err != nil {
		log.Println("Failed to send telegram video", err)
		return TelegramResponse{}, err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		log.Println("Failed to decode telegram send video response", err)
		return TelegramResponse{}, err
	}
	return telegramResp, nil
}

// SendVideoUpload multipart-uploads the raw video bytes. This raises the size
// ceiling to Telegram's 50MB bot upload limit.
func (t *TelegramClientImpl) SendVideoUpload(chatId int64, data []byte, filename, caption string) (TelegramResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("chat_id", strconv.FormatInt(chatId, 10))
	_ = writer.WriteField("caption", caption)
	_ = writer.WriteField("parse_mode", "markdown")
	_ = writer.WriteField("disable_notification", "true")

	part, err := writer.CreateFormFile("video", filename)
	if err != nil {
		return TelegramResponse{}, err
	}
	if _, err := part.Write(data); err != nil {
		return TelegramResponse{}, err
	}
	if err := writer.Close(); err != nil {
		return TelegramResponse{}, err
	}

	reqURL := fmt.Sprintf("%s/sendvideo", t.Endpoint)
	req, err := http.NewRequest("POST", reqURL, &body)
	if err != nil {
		return TelegramResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := t.uploadClient.Do(req)
	if err != nil {
		log.Println("Failed to upload telegram video", err)
		return TelegramResponse{}, err
	}
	defer resp.Body.Close()

	var telegramResp TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&telegramResp); err != nil {
		log.Println("Failed to decode telegram upload video response", err)
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
