package services

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

// Ticker structure from Binance stream
type StreamTicker struct {
	Symbol string `json:"s"` // Symbol
	Price  string `json:"c"` // Last price
	Change string `json:"P"` // Price change %
}

// Define interested symbols
var interestedSymbols = []string{
	"BGBUSDT", "CARVUSDT", "PUFFERUSDT", "BTCUSDT",
	"DOGSUSDT", "CATSUSDT", "DOGEUSDT", "SUIUSDT",
	"PEPEUSDT", "SOLUSDT",
}

// WebSocket streaming from Binance
func StartBinanceTickerStream(send chan<- StreamTicker) {
	var streams []string
	for _, symbol := range interestedSymbols {
		streams = append(streams, strings.ToLower(symbol)+"@ticker")
	}
	url := "wss://stream.binance.com:9443/stream?streams=" + strings.Join(streams, "/")

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println("WebSocket connection error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			return
		}

		var wrapper struct {
			Stream string       `json:"stream"`
			Data   StreamTicker `json:"data"`
		}

		if err := json.Unmarshal(msg, &wrapper); err != nil {
			log.Println("JSON unmarshal error:", err)
			continue
		}

		if contains(interestedSymbols, wrapper.Data.Symbol) {
			send <- wrapper.Data
		}
	}
}

// Utility to check if a symbol is in our list
func contains(list []string, item string) bool {
	for _, v := range list {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}
