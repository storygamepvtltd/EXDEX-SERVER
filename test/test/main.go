package main

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
)

// Configuration - Hardcoded for Testnet
const (
	BaseURL   = "https://testnet.binance.vision"
	WSBaseURL = "wss://stream.testnet.binance.vision:9443/ws/"
	APIKey    = "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta"
	SecretKey = "npG6JapFCZ9B3iA39ODdpdOkLOOzqoK8S1O5E7I9vgXJ1Axh28ICpcf3c2Rmfd8f"
)

type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

type ExecutionReport struct {
	EventType                string      `json:"e"`
	EventTime                int64       `json:"E"`
	Symbol                   string      `json:"s"`
	ClientOrderID            string      `json:"c"`
	Side                     string      `json:"S"`
	OrderType                string      `json:"o"`
	TimeInForce              string      `json:"f"`
	OrderQuantity            string      `json:"q"`
	OrderPrice               string      `json:"p"`
	StopPrice                string      `json:"P"`
	IcebergQuantity          string      `json:"F"`
	OrderListID              int64       `json:"g"`
	OrigClientOrderID        string      `json:"C"`
	CurrentExecutionType     string      `json:"x"`
	CurrentOrderStatus       string      `json:"X"`
	OrderRejectReason        string      `json:"r"`
	OrderID                  int64       `json:"i"`
	LastExecutedQuantity     string      `json:"l"`
	CumulativeFilledQuantity string      `json:"z"`
	LastExecutedPrice        string      `json:"L"`
	CommissionAmount         string      `json:"n"`
	CommissionAsset          string      `json:"N"`
	TransactionTime          int64       `json:"T"`
	TradeID                  int64       `json:"t"`
	OrderCreationTime        interface{} `json:"O"`
	CumulativeQuoteQty       string      `json:"Z"`
	LastQuoteQty             string      `json:"Y"`
	QuoteOrderQty            string      `json:"Q"`
}

type WSMessage struct {
	Stream string          `json:"stream,omitempty"`
	Data   ExecutionReport `json:"data,omitempty"`
	ExecutionReport
}

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

func (c *BinanceClient) sign(queryString string) string {
	h := hmac.New(sha256.New, []byte(c.SecretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

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

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response ListenKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("JSON unmarshal error: %v, body: %s", err, string(body))
	}

	return response.ListenKey, nil
}

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

func (c *BinanceClient) StreamUserData(listenKey string) error {
	wsURL := c.WSBaseURL + listenKey

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	log.Println("Connected to Binance User Data Stream")

	keepAliveTicker := time.NewTicker(30 * time.Minute)
	pingTicker := time.NewTicker(3 * time.Minute)
	defer keepAliveTicker.Stop()
	defer pingTicker.Stop()

	done := make(chan struct{})

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

			if wsMsg.EventType == "executionReport" {
				handleExecutionReport(wsMsg.ExecutionReport)
			}
		}
	}()

	for {
		select {
		case <-done:
			return nil
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Error sending close message: %v", err)
			}
			return nil
		case <-keepAliveTicker.C:
			log.Println("Sending keep-alive for listen key...")
			if err := c.KeepListenKeyAlive(listenKey); err != nil {
				log.Printf("Failed to keep listen key alive: %v", err)
			}
		case <-pingTicker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return err
			}
		}
	}
}

func handleExecutionReport(report ExecutionReport) {
	log.Printf("=== ORDER EXECUTION REPORT ===")
	log.Printf("Symbol: %s", report.Symbol)
	log.Printf("Order ID: %d", report.OrderID)
	log.Printf("Client Order ID: %s", report.ClientOrderID)
	log.Printf("Side: %s", report.Side)
	log.Printf("Order Type: %s", report.OrderType)
	log.Printf("Order Status: %s", report.CurrentOrderStatus)
	log.Printf("Execution Type: %s", report.CurrentExecutionType)

	// Detect TP/SL orders
	if report.OrderType == "STOP_LOSS" || report.OrderType == "STOP_LOSS_LIMIT" {
		log.Printf("📉 STOP LOSS Order Detected at %s", report.StopPrice)
	}
	if report.OrderType == "TAKE_PROFIT" || report.OrderType == "TAKE_PROFIT_LIMIT" {
		log.Printf("📈 TAKE PROFIT Order Detected at %s", report.StopPrice)
	}

	if report.OrderListID != -1 {
		log.Printf("🪝 OCO Group Order ID: %d", report.OrderListID)
	}

	if report.CurrentExecutionType == "TRADE" {
		log.Printf("🎉 ORDER EXECUTED!")
		log.Printf("Executed Quantity: %s", report.LastExecutedQuantity)
		log.Printf("Executed Price: %s", report.LastExecutedPrice)
		log.Printf("Cumulative Filled: %s", report.CumulativeFilledQuantity)

		if report.CommissionAmount != "" && report.CommissionAmount != "0" {
			log.Printf("Commission: %s %s", report.CommissionAmount, report.CommissionAsset)
		}

		switch report.CurrentOrderStatus {
		case "FILLED":
			log.Printf("✅ ORDER COMPLETELY FILLED")
		case "PARTIALLY_FILLED":
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

func main() {
	client := NewBinanceClient()

	log.Println("Getting listen key...")
	listenKey, err := client.GetListenKey()
	if err != nil {
		log.Fatalf("Failed to get listen key: %v", err)
	}

	log.Printf("Listen key obtained: %s", listenKey[:10]+"...")

	log.Println("Starting user data stream...")
	if err := client.StreamUserData(listenKey); err != nil {
		log.Fatalf("Stream error: %v", err)
	}
}
