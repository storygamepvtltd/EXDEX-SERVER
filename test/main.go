package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// OrderBookUpdate represents Binance depth stream response
type OrderBookUpdate struct {
	LastUpdateID int        `json:"u"`
	Bids         [][]string `json:"b"`
	Asks         [][]string `json:"a"`
}

func main() {
	symbol := "btcusdt" // example symbol
	stream := fmt.Sprintf("%s@depth", symbol)
	endpoint := url.URL{
		Scheme: "wss",
		Host:   "testnet.binance.vision",
		Path:   fmt.Sprintf("/ws/%s", stream),
	}

	log.Printf("Connecting to %s", endpoint.String())
	conn, _, err := websocket.DefaultDialer.Dial(endpoint.String(), nil)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
	}
	defer conn.Close()

	log.Println("✅ Connected to Binance WebSocket!")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var update OrderBookUpdate
		err = json.Unmarshal(message, &update)
		if err != nil {
			log.Println("JSON parse error:", err)
			continue
		}

		// Display top 1 bid and ask
		fmt.Println("📈 Live Order Book Update:")
		if len(update.Bids) > 0 {
			fmt.Printf("Bid: %s @ %s\n", update.Bids[0][1], update.Bids[0][0])
		}
		if len(update.Asks) > 0 {
			fmt.Printf("Ask: %s @ %s\n", update.Asks[0][1], update.Asks[0][0])
		}
		fmt.Println("-------------------------------")
	}
}
