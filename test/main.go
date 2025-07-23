package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL   = "https://testnet.binance.vision"
	APIKey    = "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta"
	SecretKey = "npG6JapFCZ9B3iA39ODdpdOkLOOzqoK8S1O5E7I9vgXJ1Axh28ICpcf3c2Rmfd8f"
)

type OrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  int64  `json:"transactTime"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Side          string `json:"side"`
}

type OCOOrderResponse struct {
	OrderListID       int64  `json:"orderListId"`
	ContingencyType   string `json:"contingencyType"`
	ListStatusType    string `json:"listStatusType"`
	ListOrderStatus   string `json:"listOrderStatus"`
	ListClientOrderID string `json:"listClientOrderId"`
	TransactionTime   int64  `json:"transactionTime"`
	Symbol            string `json:"symbol"`
	Orders            []struct {
		Symbol        string `json:"symbol"`
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
	} `json:"orders"`
}

// Generate HMAC SHA256 signature
func generateSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(SecretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// Make authenticated request to Binance API
func makeRequest(method, endpoint, params string) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	queryString := params + "&timestamp=" + timestamp
	signature := generateSignature(queryString)
	finalQuery := queryString + "&signature=" + signature

	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest("POST", BaseURL+endpoint, strings.NewReader(finalQuery))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest("GET", BaseURL+endpoint+"?"+finalQuery, nil)
		if err != nil {
			return nil, err
		}
	}

	req.Header.Set("X-MBX-APIKEY", APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	return body, nil
}

// 1. Market Order without TP/SL
func PlaceMarketOrder(symbol, side, quantity string) (*OrderResponse, error) {
	params := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%s", symbol, side, quantity)

	body, err := makeRequest("POST", "/api/v3/order", params)
	if err != nil {
		return nil, err
	}

	var order OrderResponse
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// 2. Market Order with TP/SL (using OCO after market order)
func PlaceMarketOrderWithTPSL(symbol, side, quantity, takeProfitPrice, stopLossPrice string) (*OrderResponse, *OCOOrderResponse, error) {
	// First place market order
	marketOrder, err := PlaceMarketOrder(symbol, side, quantity)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to place market order: %v", err)
	}

	// Determine opposite side for TP/SL orders
	oppositeSide := "SELL"
	if side == "SELL" {
		oppositeSide = "BUY"
	}

	// Place OCO order for TP/SL
	params := fmt.Sprintf("symbol=%s&side=%s&quantity=%s&price=%s&stopPrice=%s&stopLimitPrice=%s&stopLimitTimeInForce=GTC",
		symbol, oppositeSide, quantity, takeProfitPrice, stopLossPrice, stopLossPrice)

	body, err := makeRequest("POST", "/api/v3/order/oco", params)
	if err != nil {
		return marketOrder, nil, fmt.Errorf("market order placed but OCO failed: %v", err)
	}

	var ocoOrder OCOOrderResponse
	err = json.Unmarshal(body, &ocoOrder)
	if err != nil {
		return marketOrder, nil, fmt.Errorf("market order placed but OCO parsing failed: %v", err)
	}

	return marketOrder, &ocoOrder, nil
}

// 3. Limit Order without TP/SL
func PlaceLimitOrder(symbol, side, quantity, price string) (*OrderResponse, error) {
	params := fmt.Sprintf("symbol=%s&side=%s&type=LIMIT&timeInForce=GTC&quantity=%s&price=%s",
		symbol, side, quantity, price)

	body, err := makeRequest("POST", "/api/v3/order", params)
	if err != nil {
		return nil, err
	}

	var order OrderResponse
	err = json.Unmarshal(body, &order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

// 4. Limit Order with TP/SL (using OCO)
func PlaceLimitOrderWithTPSL(symbol, side, quantity, price, takeProfitPrice, stopLossPrice string) (*OrderResponse, error) {
	// For limit orders with TP/SL, we need to use a more complex approach
	// This places a limit order and sets up conditional orders

	// First place the limit order
	limitOrder, err := PlaceLimitOrder(symbol, side, quantity, price)
	if err != nil {
		return nil, fmt.Errorf("failed to place limit order: %v", err)
	}

	fmt.Printf("Limit order placed. You'll need to manually set up TP/SL after it fills, or use a trading bot.\n")
	fmt.Printf("Take Profit Price: %s, Stop Loss Price: %s\n", takeProfitPrice, stopLossPrice)

	return limitOrder, nil
}

// Utility function to get account info
func GetAccountInfo() (map[string]interface{}, error) {
	body, err := makeRequest("GET", "/api/v3/account", "")
	if err != nil {
		return nil, err
	}

	var account map[string]interface{}
	err = json.Unmarshal(body, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Utility function to get symbol info
func GetSymbolInfo(symbol string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", BaseURL+"/api/v3/exchangeInfo", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var exchangeInfo map[string]interface{}
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		return nil, err
	}

	return exchangeInfo, nil
}

func main() {
	fmt.Println("Binance Order Management System")
	fmt.Println("===============================")

	// Example usage - uncomment to test
	/*
		// Test account connection
		account, err := GetAccountInfo()
		if err != nil {
			fmt.Printf("Error getting account info: %v\n", err)
			return
		}
		fmt.Printf("Account connected successfully\n")

		// Example 1: Market Order without TP/SL
		fmt.Println("\n1. Testing Market Order without TP/SL")
		marketOrder, err := PlaceMarketOrder("BTCUSDT", "BUY", "0.001")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Market Order placed: %+v\n", marketOrder)
		}

		// Example 2: Market Order with TP/SL
		fmt.Println("\n2. Testing Market Order with TP/SL")
		marketOrderTP, ocoOrder, err := PlaceMarketOrderWithTPSL("BTCUSDT", "BUY", "0.001", "45000", "35000")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Market Order: %+v\n", marketOrderTP)
			fmt.Printf("OCO Order: %+v\n", ocoOrder)
		}

		// Example 3: Limit Order without TP/SL
		fmt.Println("\n3. Testing Limit Order without TP/SL")
		limitOrder, err := PlaceLimitOrder("BTCUSDT", "BUY", "0.001", "38000")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Limit Order placed: %+v\n", limitOrder)
		}

		// Example 4: Limit Order with TP/SL
		fmt.Println("\n4. Testing Limit Order with TP/SL")
		limitOrderTP, err := PlaceLimitOrderWithTPSL("BTCUSDT", "BUY", "0.001", "38000", "45000", "35000")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Limit Order with TP/SL setup: %+v\n", limitOrderTP)
		}
	*/

	// Interactive menu
	for {
		fmt.Println("\nSelect an option:")
		fmt.Println("1. Place Market Order (no TP/SL)")
		fmt.Println("2. Place Market Order (with TP/SL)")
		fmt.Println("3. Place Limit Order (no TP/SL)")
		fmt.Println("4. Place Limit Order (with TP/SL)")
		fmt.Println("5. Get Account Info")
		fmt.Println("6. Exit")
		fmt.Print("Enter choice (1-6): ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Print("Enter symbol (e.g., BTCUSDT): ")
			var symbol string
			fmt.Scanln(&symbol)

			fmt.Print("Enter side (BUY/SELL): ")
			var side string
			fmt.Scanln(&side)

			fmt.Print("Enter quantity: ")
			var quantity string
			fmt.Scanln(&quantity)

			order, err := PlaceMarketOrder(symbol, side, quantity)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Order placed successfully: %+v\n", order)
			}

		case 2:
			fmt.Print("Enter symbol (e.g., BTCUSDT): ")
			var symbol string
			fmt.Scanln(&symbol)

			fmt.Print("Enter side (BUY/SELL): ")
			var side string
			fmt.Scanln(&side)

			fmt.Print("Enter quantity: ")
			var quantity string
			fmt.Scanln(&quantity)

			fmt.Print("Enter take profit price: ")
			var tp string
			fmt.Scanln(&tp)

			fmt.Print("Enter stop loss price: ")
			var sl string
			fmt.Scanln(&sl)

			marketOrder, ocoOrder, err := PlaceMarketOrderWithTPSL(symbol, side, quantity, tp, sl)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Market Order: %+v\n", marketOrder)
				if ocoOrder != nil {
					fmt.Printf("OCO Order: %+v\n", ocoOrder)
				}
			}

		case 3:
			fmt.Print("Enter symbol (e.g., BTCUSDT): ")
			var symbol string
			fmt.Scanln(&symbol)

			fmt.Print("Enter side (BUY/SELL): ")
			var side string
			fmt.Scanln(&side)

			fmt.Print("Enter quantity: ")
			var quantity string
			fmt.Scanln(&quantity)

			fmt.Print("Enter price: ")
			var price string
			fmt.Scanln(&price)

			order, err := PlaceLimitOrder(symbol, side, quantity, price)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Order placed successfully: %+v\n", order)
			}

		case 4:
			fmt.Print("Enter symbol (e.g., BTCUSDT): ")
			var symbol string
			fmt.Scanln(&symbol)

			fmt.Print("Enter side (BUY/SELL): ")
			var side string
			fmt.Scanln(&side)

			fmt.Print("Enter quantity: ")
			var quantity string
			fmt.Scanln(&quantity)

			fmt.Print("Enter limit price: ")
			var price string
			fmt.Scanln(&price)

			fmt.Print("Enter take profit price: ")
			var tp string
			fmt.Scanln(&tp)

			fmt.Print("Enter stop loss price: ")
			var sl string
			fmt.Scanln(&sl)

			order, err := PlaceLimitOrderWithTPSL(symbol, side, quantity, price, tp, sl)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Order placed successfully: %+v\n", order)
			}

		case 5:
			account, err := GetAccountInfo()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Account Info: %+v\n", account)
			}

		case 6:
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice. Please select 1-6.")
		}
	}
}
