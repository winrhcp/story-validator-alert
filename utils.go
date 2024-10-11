package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	TelegramBotToken string `json:"telegramBotToken"`
	TelegramChatID   string `json:"telegramChatID"`
	RPCEndpoint      string `json:"rpcEndpoint"`
}

type StatusResponse struct {
	Result struct {
		SyncInfo struct {
			LatestBlockHeight int64 `json:"latest_block_height"`
		} `json:"sync_info"`
	} `json:"result"`
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(filename string) Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	var config Config
	json.Unmarshal(data, &config)
	return config
}

// CheckNodeStatus checks the status of the node and sends alerts if needed
func CheckNodeStatus(endpoint string, bot *tgbotapi.BotAPI, chatID string) {
	resp, err := http.Get(endpoint)
	if err != nil {
		SendAlert(bot, chatID, "Node is down! Error: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		SendAlert(bot, chatID, "Node is not responding! Status code: "+fmt.Sprintf("%d", resp.StatusCode))
		return
	}

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		SendAlert(bot, chatID, "Failed to decode response: "+err.Error())
		return
	}

	// Example condition: alert if the latest block height is too low (indicating potential downtime)
	if status.Result.SyncInfo.LatestBlockHeight < 100000 {
		SendAlert(bot, chatID, fmt.Sprintf("Warning: Latest block height is low (%d).", status.Result.SyncInfo.LatestBlockHeight))
	}
}

// CheckNodePeers retrieves and logs the list of peers
func CheckNodePeers() {
	cmd := exec.Command("curl", "localhost:26657/net_info")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Printf("Error checking peers: %v", err)
		return
	}

	var result struct {
		Result struct {
			Peers []struct {
				NodeInfo struct {
					Moniker string `json:"moniker"`
				} `json:"node_info"`
			} `json:"peers"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		log.Printf("Failed to decode peers response: %v", err)
		return
	}

	log.Println("Connected peers:")
	for _, peer := range result.Result.Peers {
		log.Println(peer.NodeInfo.Moniker)
	}
}

// CheckNodeHealth checks the health of the node
func CheckNodeHealth() bool {
	resp, err := http.Get("http://localhost:26657/health")
	if err != nil {
		log.Printf("Error checking health: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Node is unhealthy!")
		return false
	}

	// Health response is empty JSON object for healthy node
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "{}" {
		log.Println("Node is healthy.")
		return true
	}

	log.Println("Unexpected health response:", string(body))
	return false
}

// SendAlert sends an alert message to the specified Telegram chat
func SendAlert(bot *tgbotapi.BotAPI, chatID string, message string) {
	msg := tgbotapi.NewMessageToChannel(chatID, message)
	bot.Send(msg)
}
