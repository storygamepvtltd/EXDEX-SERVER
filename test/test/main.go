package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Replace with your actual listenKey from REST API
const (
	listenKey      = "tWCiWsjhuGAAKL2MPXMKgVkZ3mVbdXc0lEh75WqBJGzA7UHGrGxPItlf0l2d"
	apiKey         = "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta"
	orderIDToTrack = 2041308   // Replace with your actual orderId
	symbolToTrack  = "LTCUSDT" // Uppercase symbol
)

func main() {
	url := "wss://stream.binance.com:9443/ws/" + listenKey

	headers := http.Header{}
	headers.Add("X-MBX-APIKEY", apiKey)

	c, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		log.Fatalf("WebSocket connection error: %v", err)
	}
	defer c.Close()

	log.Println("Connected to Binance WebSocket!")

	// Set ping to keep alive
	go func() {
		for {
			err := c.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Println("Ping error:", err)
				return
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			log.Println("JSON error:", err)
			continue
		}

		if event["e"] == "executionReport" {
			orderID := int64(event["i"].(float64))
			symbol := event["s"].(string)

			if orderID == orderIDToTrack && symbol == symbolToTrack {
				status := event["X"]
				log.Printf("Order Update: OrderID %d, Symbol %s, Status: %s", orderID, symbol, status)
			}
		}
	}
}
