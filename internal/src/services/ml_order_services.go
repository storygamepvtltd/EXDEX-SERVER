package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	models "exdex/internal/src/model"
	"exdex/internal/src/repository"
	database "exdex/server/databases"
)

func generateSignature(query string) string {
	h := hmac.New(sha256.New, []byte("npG6JapFCZ9B3iA39ODdpdOkLOOzqoK8S1O5E7I9vgXJ1Axh28ICpcf3c2Rmfd8f"))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

func makeRequest(method, endpoint, params string) ([]byte, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	query := params + "&timestamp=" + timestamp
	signature := generateSignature(query)
	finalQuery := query + "&signature=" + signature

	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest("POST", "https://testnet.binance.vision"+endpoint, strings.NewReader(finalQuery))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest("GET", "https://testnet.binance.vision"+endpoint+"?"+finalQuery, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-MBX-APIKEY", "b5J1Wa8FEi7TVeeO6jLOsWBdEdmPsGmwB4gxJXT0F4YKlQx3VXmTtSeVn4Wuhwta")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, errors.New(string(body))
	}

	return body, nil
}

func PlaceMarketOrder(symbol, side, quantity, userID string) (*models.OrderResponse, error) {
	params := fmt.Sprintf("symbol=%s&side=%s&type=MARKET&quantity=%s", symbol, side, quantity)
	body, err := makeRequest("POST", "/api/v3/order", params)
	if err != nil {
		return nil, err
	}
	var order models.OrderResponse
	json.Unmarshal(body, &order)
	err = storeOrderToDB(&order, side, "MARKET", quantity, "", "", "", userID)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func PlaceMarketOrderWithTPSL(symbol, side, quantity, tp, sl, userID string) (*models.OrderResponse, *models.OCOOrderResponse, error) {
	order, err := PlaceMarketOrder(symbol, side, quantity, userID)
	if err != nil {
		return nil, nil, err
	}

	opposite := "SELL"
	if side == "SELL" {
		opposite = "BUY"
	}

	params := fmt.Sprintf("symbol=%s&side=%s&quantity=%s&price=%s&stopPrice=%s&stopLimitPrice=%s&stopLimitTimeInForce=GTC",
		symbol, opposite, quantity, tp, sl, sl)

	body, err := makeRequest("POST", "/api/v3/order/oco", params)
	if err != nil {
		return order, nil, err
	}

	var oco models.OCOOrderResponse
	json.Unmarshal(body, &oco)

	// order, err := PlaceMarketOrder(symbol, side, quantity)
	err = storeOrderToDB(order, side, "MARKET", quantity, "", tp, sl, userID)
	if err != nil {
		return order, nil, err
	}
	return order, &oco, nil
}

func PlaceLimitOrder(symbol, side, quantity, price, userID string) (*models.OrderResponse, error) {
	params := fmt.Sprintf("symbol=%s&side=%s&type=LIMIT&timeInForce=GTC&quantity=%s&price=%s",
		symbol, side, quantity, price)

	body, err := makeRequest("POST", "/api/v3/order", params)
	if err != nil {
		return nil, err
	}

	var order models.OrderResponse
	json.Unmarshal(body, &order)

	err = storeOrderToDB(&order, side, "LIMIT", quantity, price, "", "", userID)

	if err != nil {
		return &order, nil
	}
	return &order, nil
}

func PlaceLimitOrderWithTPSL(symbol, side, quantity, price, tp, sl, userID string) (*models.OrderResponse, error) {
	order, err := PlaceLimitOrder(symbol, side, quantity, price, userID)
	if err != nil {
		return nil, err
	}

	err = storeOrderToDB(order, side, "LIMIT", quantity, price, tp, sl, userID)

	if err != nil {
		return order, nil
	}
	// TP/SL is manual or external logic after limit fill.
	return order, nil
}
func GetMyOrders(userID string, page int, limit int) ([]models.OrderRecord, int64, error) {
	var orders []models.OrderRecord

	// Convert userID to ObjectID if needed (uncomment below if user_id is ObjectID in MongoDB)
	// objID, err := primitive.ObjectIDFromHex(userID)
	// if err != nil {
	//     return nil, 0, err
	// }

	filter := bson.M{
		"user_id": userID, // use objID here if you uncomment above
	}

	skip := (page - 1) * limit

	opts := options.Find()
	opts.SetSkip(int64(skip))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	collection := database.DB.Collection("orders")

	// Get total count of documents matching the filter
	total, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return nil, 0, err
	}

	// Find paginated results
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	// Decode results into the orders slice
	if err := cursor.All(context.TODO(), &orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func storeOrderToDB(order *models.OrderResponse, side, orderType, quantity, price, tp, sl, userID string) error {
	record := models.OrderRecord{
		Symbol:          order.Symbol,
		Side:            side,
		Type:            orderType,
		Quantity:        quantity,
		Price:           price,
		TakeProfitPrice: tp,
		StopLossPrice:   sl,
		UserID:          userID,
		OrderID:         order.OrderID,
		ClientOrderID:   order.ClientOrderID,
		Status:          order.Status,
		CreatedAt:       primitive.NewDateTimeFromTime(time.Now()),
	}

	err := repository.IRepo.Insert("orders", &record)
	if err != nil {
		return err
	}

	return nil
}
