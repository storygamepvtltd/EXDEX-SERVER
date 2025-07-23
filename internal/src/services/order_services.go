package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
		Sort:       "createdAt",
		SortOrder:  -1,
		Limit:      limit,
		Offset:     offset,
		Conditions: bson.M{"user_id": uId},
	}

	var items []models.PlacedOrder
	err := repository.IRepo.GetAllByFiltter("orders", &items, filter.Conditions, filter)
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

	if err != nil {
		return models.PlacedOrder{}, fmt.Errorf("order placement failed: %v", err)
	}

	placedOrder, err := storeOrderInMongo(req, res, uId)
	if err != nil {
		return placedOrder, fmt.Errorf("failed to store order: %v", err)
	}

	return placedOrder, nil
}

func marketOrderWithTPSL(req models.OrderRequest) (string, error) {
	// Place main market order
	// orderResp, err := simpleOrder(req)
	// if err != nil {
	// 	return "", fmt.Errorf("main order failed: %v", err)
	// }

	if !req.WithTPSL {
		return "orderResp", nil
	}

	// Get current price
	currentPrice, err := getLastPrice(req.Symbol)
	if err != nil {
		return "orderResp", fmt.Errorf("failed to get last price: %v", err)
	}

	priceF, err := strconv.ParseFloat(currentPrice, 64)
	if err != nil {
		return "orderResp", fmt.Errorf("failed to parse price: %v", err)
	}

	var tpPrice, slPrice float64
	if req.UseAbs {
		tpPrice = req.TPPrice
		slPrice = req.SLPrice
	} else {
		if req.Side == "BUY" {
			tpPrice = priceF * (1 + req.TPMul)
			slPrice = priceF * (1 - req.SLMul)
		} else {
			tpPrice = priceF * (1 - req.TPMul)
			slPrice = priceF * (1 + req.SLMul)
		}
	}

	// Validate prices
	if req.Side == "BUY" {
		if tpPrice <= priceF || slPrice >= priceF {
			return "orderResp", errors.New("for BUY orders, TP must be above current price and SL must be below")
		}
	} else {
		if tpPrice >= priceF || slPrice <= priceF {
			return "orderResp", errors.New("for SELL orders, TP must be below current price and SL must be above")
		}
	}

	// Place OCO order
	ocoResp, err := placeOCO(req.Symbol, oppositeSide(req.Side), req.Quantity, tpPrice, slPrice)
	if err != nil {
		return fmt.Sprintf("%s (OCO failed: %v)", "orderResp", err), nil
	}

	return fmt.Sprintf("Market Order: %s\nOCO TP/SL: %s", "orderResp", ocoResp), nil
}

func placeOCO(symbol, side, qty string, tpPrice, slPrice float64) (string, error) {
	endpoint := "/api/v3/order/oco"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Calculate stop limit price with small buffer
	stopLimitPrice := slPrice
	if side == "SELL" {
		stopLimitPrice = slPrice * 1.001
	} else {
		stopLimitPrice = slPrice * 0.999
	}

	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("side", side)
	data.Set("quantity", qty)
	// data.Set("a", "aaaa")
	data.Set("price", fmt.Sprintf("%.8f", tpPrice))
	data.Set("stopPrice", fmt.Sprintf("%.8f", slPrice))
	data.Set("stopLimitPrice", fmt.Sprintf("%.8f", stopLimitPrice))
	data.Set("stopLimitTimeInForce", "GTC")
	data.Set("timestamp", timestamp)
	data.Set("recvWindow", "5000")

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	return sendRequest("POST", fullURL)
}

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

func placeOCOAbsolute(req models.OrderRequest) (string, error) {
	return sendOCO(req.Side, req.Quantity,
		fmt.Sprintf("%.8f", req.TPPrice),
		fmt.Sprintf("%.8f", req.SLPrice),
		req.Symbol)
}

func placeOCODynamic(req models.OrderRequest) (string, error) {
	price := getPriceFloat(req.Symbol)
	var tp, sl float64

	if req.Side == "BUY" {
		tp = price * (1 + req.TPMul)
		sl = price * (1 - req.SLMul)
	} else {
		tp = price * (1 - req.TPMul)
		sl = price * (1 + req.SLMul)
	}

	return sendOCO(req.Side, req.Quantity,
		fmt.Sprintf("%.8f", tp),
		fmt.Sprintf("%.8f", sl),
		req.Symbol)
}
func sendOCO(side, qty, tp, sl, symbol string) (string, error) {
	endpoint := "/api/v3/order/oco"
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Convert TP and SL to float
	tpFloat, _ := strconv.ParseFloat(tp, 64)
	slFloat, _ := strconv.ParseFloat(sl, 64)

	// Check price relationship for SELL OCO (after BUY entry)
	if !(slFloat < tpFloat) {
		return "", fmt.Errorf("OCO error: stop-loss price %.2f must be less than take-profit price %.2f", slFloat, tpFloat)
	}

	// Calculate stopLimitPrice slightly below SL to ensure trigger
	stopLimitPrice := slFloat * 0.999

	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("side", side)
	data.Set("quantity", qty)
	data.Set("price", fmt.Sprintf("%.2f", tpFloat))                 // Take Profit
	data.Set("stopPrice", fmt.Sprintf("%.2f", slFloat))             // Stop Trigger
	data.Set("stopLimitPrice", fmt.Sprintf("%.8f", stopLimitPrice)) // Actual Stop-Limit Sell Price
	data.Set("stopLimitTimeInForce", "GTC")
	data.Set("recvWindow", "5000")
	data.Set("timestamp", timestamp)

	sign := signQuery(data.Encode())
	data.Set("signature", sign)

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, data.Encode())
	return sendRequest("POST", fullURL)
}

func getLastPrice(symbol string) (string, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", viper.GetString("binance.baseURL"), symbol)
	body, err := sendRequest("GET", url)
	if err != nil {
		return "", err
	}

	var res struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return "", err
	}
	return res.Price, nil
}

func getPriceFloat(symbol string) float64 {
	priceStr, err := getLastPrice(symbol)
	if err != nil {
		return 0
	}
	price, _ := strconv.ParseFloat(priceStr, 64)
	return price
}

func fetchPrice(mult float64, symbol string) string {
	price := getPriceFloat(symbol)
	return fmt.Sprintf("%.8f", price*mult)
}

func signQuery(query string) string {
	mac := hmac.New(sha256.New, []byte(viper.GetString("binance.apiSecret")))
	mac.Write([]byte(query))
	return hex.EncodeToString(mac.Sum(nil))
}

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

func storeOrderInMongo(req models.OrderRequest, response, userId string) (models.PlacedOrder, error) {
	var parsedResp models.PlacedOrder
	err := json.Unmarshal([]byte(response), &parsedResp)
	if err != nil {
		parsedResp = models.PlacedOrder{
			// Response: response,
		}
	}

	parsedResp.UserID = userId
	parsedResp.Symbol = req.Symbol
	parsedResp.Side = req.Side
	// parsedResp.OrderType = req.OrderType
	// parsedResp.Quantity = req.Quantity
	// parsedResp.WithTPSL = req.WithTPSL
	// parsedResp.UseAbs = req.UseAbs
	// parsedResp.TPPrice = req.TPPrice
	// parsedResp.SLPrice = req.SLPrice
	// parsedResp.TPMul = req.TPMul
	// parsedResp.SLMul = req.SLMul
	// parsedResp.CreatedAt = time.Now()
	// parsedResp.Response = response

	var result map[string]interface{}
	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
	}

	// Extract clientOrderId
	if clientOrderID, ok := result["clientOrderId"]; ok {
		fmt.Println("clientOrderId:", clientOrderID)
	} else {
		fmt.Println("clientOrderId not found")
	}

	// helper.ExtractSection()
	err = repository.IRepo.Create("orders", &parsedResp)
	if err != nil {
		return parsedResp, fmt.Errorf("failed to save order in MongoDB: %v", err)
	}

	return parsedResp, nil
}

func oppositeSide(side string) string {
	if side == "BUY" {
		return "SELL"
	}
	return "BUY"
}

func (os OrderSerices) GetSymbolPrice(symbol string) (string, error) {
	return fetchPrice(1.00, symbol), nil
}

func (os OrderSerices) CancelOCOOrder(orderListId int64, symbol string) (string, error) {
	endpoint := "/api/v3/orderList"
	params := url.Values{}
	params.Add("orderListId", fmt.Sprintf("%d", orderListId))
	params.Add("symbol", symbol)
	params.Add("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Add("recvWindow", "5000")

	query := params.Encode()
	signature := utils.Sign(query, viper.GetString("binance.apiSecret"))
	query += "&signature=" + signature

	fullURL := fmt.Sprintf("%s%s?%s", viper.GetString("binance.baseURL"), endpoint, query)

	req, err := http.NewRequest("DELETE", fullURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-MBX-APIKEY", viper.GetString("binance.apiKey"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Binance API error: %s", body)
	}

	return string(body), nil
}
