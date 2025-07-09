package main

import (
	"github.com/gin-gonic/gin"
)

func telegramWebhookServiceReceive(payload string) any {
	return gin.H{"status": "received", "payload": payload}
}
