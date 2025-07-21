package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"

	models "exdex/internal/src/model"
	"exdex/internal/src/repository"
	"exdex/server/utils"
)

type OrderSerices struct{}

func (os OrderSerices) GetHistory(uId string, limit, offset int) ([]models.PlacedOrder, error) {

	filter := models.Filter{
		Sort:       "createdAt", // example sort field
		SortOrder:  -1,          // -1 = descending, 1 = ascending
		Limit:      limit,
		Offset:     offset,
		Conditions: bson.M{"user_id": uId}, // 👈 your filter condition
	}

	query := filter.Conditions // using the embedded Conditions field

	var items []models.PlacedOrder // Replace with your actual model type

	err := repository.IRepo.GetAllByFiltter("orders", &items, query, filter)
	if err != nil {
		return items, err
	}

	return items, nil

}
func (os OrderSerices) PlaceOrder(req models.OrderRequest, uId string) (models.PlacedOrder, error) {
	var res string
	var err error

	if req.WithTPSL {
		if req.OrderType == "LIMIT" {
			fmt.Println("dd", req)
			if req.UseAbs {
				res, err = placeOCOAbsolute(req)
			} else {
				res, err = placeOCODynamic(req)
			}
		} else if req.OrderType == "MARKET" {

			res, err = marketOrderWithTPSL(req)
		} else {
			res, err = simpleOrder(req)
		}
	} else {
		res, err = simpleOrder(req)
	}

	// Save to MongoDB
	placedOrder, err := storeOrderInMongo(req, res, uId)
	if err != nil {
		return placedOrder, err
	}

	return placedOrder, nil
}

// Basic order (MARKET / LIMIT without TP/SL)
func simpleOrder(req models.OrderRequest) (string, error) {
	endpoint := "/api/v3/order"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	data := url.Values{}
	data.Set("symbol", req.Symbol)
	data.Set("side", req.Side)
	data.Set("type", req.OrderType)
	data.Set("quantity", req.Quantity)
	data.Set("timestamp", timestamp)

	if req.OrderType == "LIMIT" {
		data.Set("price", fetchPrice(1.00, req.Symbol))
		data.Set("timeInForce", "GTC")
	}

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	return sendRequest("POST", fullURL)
}

// MARKET + TP/SL (Absolute or Dynamic)
func marketOrderWithTPSL(req models.OrderRequest) (string, error) {
	// Step 1: Market order
	res, err := simpleOrder(req)
	if err != nil {
		return res, err
	}

	// Step 2: Calculate TP/SL dynamically if needed
	if !req.UseAbs {

		price := getPriceFloat(req.Symbol)
		req.TPPrice = price * req.TPMul
		req.SLPrice = price * req.SLMul
	}

	// Step 3: SL Order
	err = placeStopLossOrder(req)
	if err != nil {
		return res, fmt.Errorf("market order placed, failed SL: %v", err)
	}

	// Step 4: TP Order
	err = placeTakeProfitOrder(req)
	if err != nil {
		return res, fmt.Errorf("market + SL placed, failed TP: %v", err)
	}

	return res, nil
}

// LIMIT + TP/SL - Absolute
func placeOCOAbsolute(req models.OrderRequest) (string, error) {
	return sendOCO(req.Side, req.Quantity,
		fmt.Sprintf("%.2f", req.TPPrice),
		fmt.Sprintf("%.2f", req.SLPrice),
		req.Symbol)
}

// LIMIT + TP/SL - Dynamic
func placeOCODynamic(req models.OrderRequest) (string, error) {
	price := getPriceFloat(req.Symbol)
	tp := price * req.TPMul
	sl := price * req.SLMul
	return sendOCO(req.Side, req.Quantity,
		fmt.Sprintf("%.2f", tp),
		fmt.Sprintf("%.2f", sl),
		req.Symbol)
}

// Send OCO Order
func sendOCO(side, qty, tp, sl, symbol string) (string, error) {
	endpoint := "/api/v3/order/oco"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("side", side)
	data.Set("quantity", qty)
	data.Set("price", tp)
	data.Set("stopPrice", sl)
	data.Set("stopLimitPrice", sl)
	data.Set("stopLimitTimeInForce", "GTC")
	data.Set("recvWindow", "5000")
	data.Set("timestamp", timestamp)

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	return sendRequest("POST", fullURL)
}

// Stop Loss for MARKET
func placeStopLossOrder(req models.OrderRequest) error {
	endpoint := "/api/v3/order"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	data := url.Values{}
	data.Set("symbol", req.Symbol)
	data.Set("side", oppositeSide(req.Side))
	data.Set("type", "STOP_LOSS")
	data.Set("stopPrice", fmt.Sprintf("%.2f", req.SLPrice))
	data.Set("quantity", req.Quantity)
	data.Set("recvWindow", "5000")
	data.Set("timestamp", timestamp)

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	_, err := sendRequest("POST", fullURL)
	return err
}

// Take Profit for MARKET
func placeTakeProfitOrder(req models.OrderRequest) error {
	endpoint := "/api/v3/order"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	data := url.Values{}
	data.Set("symbol", req.Symbol)
	data.Set("side", oppositeSide(req.Side))
	data.Set("type", "TAKE_PROFIT")
	data.Set("stopPrice", fmt.Sprintf("%.2f", req.TPPrice))
	data.Set("quantity", req.Quantity)
	data.Set("recvWindow", "5000")
	data.Set("timestamp", timestamp)

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	_, err := sendRequest("POST", fullURL)
	return err
}

// Fetch current price as string
func fetchPrice(mult float64, symbol string) string {
	resp, err := http.Get(viper.GetString("binance.baseURL") + "/api/v3/ticker/price?symbol=" + symbol)
	if err != nil {
		return "0"
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]string
	json.Unmarshal(body, &data)
	price, _ := strconv.ParseFloat(data["price"], 64)
	return fmt.Sprintf("%.2f", price*mult)
}

// Fetch current price as float
func getPriceFloat(symbol string) float64 {
	resp, _ := http.Get(viper.GetString("binance.baseURL") + "/api/v3/ticker/price?symbol=" + symbol)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]string
	json.Unmarshal(body, &data)
	p, _ := strconv.ParseFloat(data["price"], 64)
	return p
}

// Signature HMAC
func signQuery(query string) string {
	mac := hmac.New(sha256.New, []byte(viper.GetString("binance.apiSecret")))
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

// Send signed HTTP request
func sendRequest(method, url string) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(nil))
	if err != nil {
		return "", err
	}
	req.Header.Set("X-MBX-APIKEY", viper.GetString("binance.apiKey"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// Save order to MongoDB
func storeOrderInMongo(req models.OrderRequest, response, userId string) (models.PlacedOrder, error) {
	var parsedResp models.PlacedOrder
	err := json.Unmarshal([]byte(response), &parsedResp)
	if err != nil {
		parsedResp = models.PlacedOrder{
			Response: response,
		}
	}

	parsedResp.UserID = userId
	parsedResp.Symbol = req.Symbol
	parsedResp.Side = req.Side
	parsedResp.OrderType = req.OrderType
	parsedResp.Quantity = req.Quantity
	parsedResp.WithTPSL = req.WithTPSL
	parsedResp.UseAbs = req.UseAbs
	parsedResp.TPPrice = req.TPPrice
	parsedResp.SLPrice = req.SLPrice
	parsedResp.TPMul = req.TPMul
	parsedResp.SLMul = req.SLMul
	parsedResp.CreatedAt = time.Now()
	parsedResp.Response = response

	err = repository.IRepo.Create("orders", &parsedResp)
	if err != nil {
		return parsedResp, fmt.Errorf("failed to save order in MongoDB: %v", err)
	}

	return parsedResp, nil
}

// Reverse BUY to SELL, or vice versa
func oppositeSide(side string) string {
	if side == "BUY" {
		return "SELL"
	}
	return "BUY"
}

// Get current price (for API call)
func (os OrderSerices) GetSymbolPrice(symbol string) (string, error) {
	return fetchPrice(1.00, symbol), nil
}

const (
	apiKey    = "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta"
	apiSecret = "npG6JapFCZ9B3iA39ODdpdOkLOOzqoK8S1O5E7I9vgXJ1Axh28ICpcf3c2Rmfd8f"
	baseURL   = "https://testnet.binance.vision"
)

func (os OrderSerices) CancelOCOOrder(orderListId int64, symbol string) (string, error) {
	endpoint := "/api/v3/orderList"
	params := url.Values{}
	params.Add("orderListId", fmt.Sprintf("%d", orderListId))
	params.Add("symbol", symbol)
	params.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Add("recvWindow", "5000")

	query := params.Encode()
	signature := utils.Sign(query, apiSecret)
	query += "&signature=" + signature

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, query)

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("❌ Binance API error: %s", body)
	}

	return string(body), nil
}
