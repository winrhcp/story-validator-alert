// main.go
package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Load configuration
	config := LoadConfig("config.json")

	// Create Telegram bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// Check node status
		CheckNodeStatus(config.RPCEndpoint, bot, config.TelegramChatID)

		// Check node peers
		CheckNodePeers()

		// Check node health
		if !CheckNodeHealth() {
			SendAlert(bot, config.TelegramChatID, "Node is unhealthy!")
		}

		time.Sleep(60 * time.Second) // Check every minute
	}
}
