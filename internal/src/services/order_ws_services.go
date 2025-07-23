package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"

	"exdex/internal/src/repository"
)

const (
	// Testnet URLs (hardcoded)
	BaseURL   = "https://testnet.binance.vision"
	WSBaseURL = "wss://stream.testnet.binance.vision:9443/ws/"

	// Testnet API Keys (hardcoded - FOR TESTING ONLY!)
	APIKey    = "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta"
	SecretKey = "npG6JapFCZ9B3iA39ODdpdOkLOOzqoK8S1O5E7I9vgXJ1Axh28ICpcf3c2Rmfd8f"
)

func OrderWSHandler() {
	client := NewBinanceClient()

	// Get listen key
	log.Println("Getting listen key...")
	listenKey, err := client.GetListenKey()
	if err != nil {
		log.Fatalf("Failed to get listen key: %v", err)
	}

	log.Printf("Listen key obtained: %s", listenKey[:10]+"...")

	// Start streaming user data
	log.Println("Starting user data stream...")
	if err := client.StreamUserData(listenKey); err != nil {
		log.Fatalf("Stream error: %v", err)
	}
}

// Structs for API responses
type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

type ExecutionReport struct {
	EventType                string      `json:"e"` // Event type
	EventTime                int64       `json:"E"` // Event time
	Symbol                   string      `json:"s"` // Symbol
	ClientOrderID            string      `json:"c"` // Client order ID
	Side                     string      `json:"S"` // Side
	OrderType                string      `json:"o"` // Order type
	TimeInForce              string      `json:"f"` // Time in force
	OrderQuantity            string      `json:"q"` // Order quantity
	OrderPrice               string      `json:"p"` // Order price
	StopPrice                string      `json:"P"` // Stop price
	IcebergQuantity          string      `json:"F"` // Iceberg quantity
	OrderListID              int64       `json:"g"` // OrderListId
	OrigClientOrderID        string      `json:"C"` // Original client order ID
	CurrentExecutionType     string      `json:"x"` // Current execution type
	CurrentOrderStatus       string      `json:"X"` // Current order status
	OrderRejectReason        string      `json:"r"` // Order reject reason
	OrderID                  int64       `json:"i"` // Order ID
	LastExecutedQuantity     string      `json:"l"` // Last executed quantity
	CumulativeFilledQuantity string      `json:"z"` // Cumulative filled quantity
	LastExecutedPrice        string      `json:"L"` // Last executed price
	CommissionAmount         string      `json:"n"` // Commission amount
	CommissionAsset          string      `json:"N"` // Commission asset
	TransactionTime          int64       `json:"T"` // Transaction time
	TradeID                  int64       `json:"t"` // Trade ID
	OrderCreationTime        interface{} `json:"O"` // Order creation time (can be int64 or array)
	CumulativeQuoteQty       string      `json:"Z"` // Cumulative quote asset transacted quantity
	LastQuoteQty             string      `json:"Y"` // Last quote asset transacted quantity
	QuoteOrderQty            string      `json:"Q"` // Quote Order Qty
}

type WSMessage struct {
	Stream string          `json:"stream,omitempty"`
	Data   ExecutionReport `json:"data,omitempty"`
	ExecutionReport
}

// BinanceClient handles API interactions
type BinanceClient struct {
	APIKey    string
	SecretKey string
	BaseURL   string
	WSBaseURL string
}

func NewBinanceClient() *BinanceClient {
	log.Println("🧪 Running on Binance Testnet - Hardcoded credentials")
	return &BinanceClient{
		APIKey:    APIKey,
		SecretKey: SecretKey,
		BaseURL:   BaseURL,
		WSBaseURL: WSBaseURL,
	}
}

// Generate signature for authenticated requests
func (c *BinanceClient) sign(queryString string) string {
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// Get listen key for user data stream
func (c *BinanceClient) GetListenKey() (string, error) {
	endpoint := "/api/v3/userDataStream"
	url := c.BaseURL + endpoint

	log.Printf("Requesting listen key from: %s", url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-MBX-APIKEY", c.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Printf("Response status: %d", resp.StatusCode)
	log.Printf("Response body: %s", string(body))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response ListenKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("JSON unmarshal error: %v, body: %s", err, string(body))
	}

	return response.ListenKey, nil
}

// Keep listen key alive (call every 30 minutes)
func (c *BinanceClient) KeepListenKeyAlive(listenKey string) error {
	endpoint := "/api/v3/userDataStream"

	data := url.Values{}
	data.Set("listenKey", listenKey)

	req, err := http.NewRequest("PUT", c.BaseURL+endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", c.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Connect to user data stream and handle order execution events
func (c *BinanceClient) StreamUserData(listenKey string) error {
	wsURL := c.WSBaseURL + listenKey

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to Binance User Data Stream")

	// Setup keep-alive ticker (every 30 minutes)
	keepAliveTicker := time.NewTicker(30 * time.Minute)
	defer keepAliveTicker.Stop()

	// Setup ping ticker (every 3 minutes)
	pingTicker := time.NewTicker(3 * time.Minute)
	defer pingTicker.Stop()

	done := make(chan struct{})

	// Read messages
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(message, &wsMsg); err != nil {
				log.Printf("JSON unmarshal error: %v", err)
				continue
			}

			// Handle execution report (order updates)
			if wsMsg.EventType == "executionReport" {
				handleExecutionReport(wsMsg.ExecutionReport)
			}
		}
	}()

	// Main loop
	for {
		select {
		case <-done:
			return nil
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")

			// Send close message
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Error sending close message: %v", err)
				return err
			}

			// Wait for close or timeout
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		case <-keepAliveTicker.C:
			log.Println("Sending keep-alive for listen key...")
			if err := c.KeepListenKeyAlive(listenKey); err != nil {
				log.Printf("Failed to keep listen key alive: %v", err)
			}
		case <-pingTicker.C:
			// Send ping to keep connection alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return err
			}
		}
	}
}

// Handle order execution events
func handleExecutionReport(report ExecutionReport) {
	log.Printf("=== ORDER EXECUTION REPORT ===")
	log.Printf("Symbol: %s", report.Symbol)
	log.Printf("Order ID: %d", report.OrderID)
	log.Printf("Client Order ID: %s", report.ClientOrderID)
	log.Printf("Side: %s", report.Side)
	log.Printf("Order Type: %s", report.OrderType)
	log.Printf("Order Status: %s", report.CurrentOrderStatus)
	log.Printf("Execution Type: %s", report.CurrentExecutionType)

	// Check if order was filled (partially or completely)
	if report.CurrentExecutionType == "TRADE" {
		log.Printf("🎉 ORDER EXECUTED!")
		log.Printf("Executed Quantity: %s", report.LastExecutedQuantity)
		log.Printf("Executed Price: %s", report.LastExecutedPrice)
		log.Printf("Cumulative Filled: %s", report.CumulativeFilledQuantity)
		log.Printf("Total Order Quantity: %s", report.OrderQuantity)

		if report.CommissionAmount != "" && report.CommissionAmount != "0" {
			log.Printf("Commission: %s %s", report.CommissionAmount, report.CommissionAsset)
		}

		// Check if fully filled
		if report.CurrentOrderStatus == "FILLED" {

			filter := bson.M{
				"order_id": report.OrderID,
			}
			update := bson.M{
				"$set": bson.M{
					"status": "filled",
				},
			}

			err := repository.IRepo.UpdateOne("orders", filter, update, false)
			if err != nil {
				log.Printf("❌ Failed to update order: %v", err)
			} else {
				log.Println("✅ Order status updated to 'filled'")
			}

			log.Printf("✅ ORDER COMPLETELY FILLED")
		} else if report.CurrentOrderStatus == "PARTIALLY_FILLED" {
			log.Printf("⚡ ORDER PARTIALLY FILLED")
		}
	} else if report.CurrentOrderStatus == "NEW" {
		log.Printf("📝 New order placed")
	} else if report.CurrentOrderStatus == "CANCELED" {
		log.Printf("❌ Order canceled")
	} else if report.CurrentOrderStatus == "REJECTED" {
		log.Printf("🚫 Order rejected: %s", report.OrderRejectReason)
	}

	log.Printf("Transaction Time: %s", time.Unix(report.TransactionTime/1000, 0).Format("2006-01-02 15:04:05"))
	log.Printf("==============================\n")
}
