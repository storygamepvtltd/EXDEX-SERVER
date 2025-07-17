package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Structure for Binance 24hr ticker data
type Ticker24hr struct {
	Symbol             string `json:"symbol"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
}

func main() {
	// List of symbols you want
	symbols := []string{"BGBUSDT", "CARVUSDT", "PUFFERUSDT", "BTCUSDT", "DOGSUSDT", "CATSUSDT", "DOGEUSDT", "SUIUSDT", "PEPEUSDT", "SOLUSDT"}

	// Binance API URL
	url := "https://api.binance.com/api/v3/ticker/24hr"

	// Send HTTP request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()

	// Decode JSON
	var tickers []Ticker24hr
	err = json.NewDecoder(resp.Body).Decode(&tickers)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Loop and print only needed symbols
	for _, ticker := range tickers {
		if contains(symbols, ticker.Symbol) {
			fmt.Printf("%s => %s%% | Last Price: %s\n", ticker.Symbol, ticker.PriceChangePercent, ticker.LastPrice)
		}
	}
}

// Helper function to check if a symbol is in list
func contains(list []string, item string) bool {
	for _, v := range list {
		if strings.EqualFold(v, item) {
			return true
		}
	}
	return false
}
