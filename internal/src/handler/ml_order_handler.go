package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"exdex/internal/src/services"
	"exdex/server/constant"
	response "exdex/server/responses"
)

func GetOrderAllHistory(c *fiber.Ctx) error {

	userID, err := getUserID(c)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	// skip := (page - 1) * limit
	data, total, err := services.GetMyOrders(userID, page, limit)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "successfully", fiber.Map{
		"data":  data,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// getUserID extracts userID from context
func getUserID(c *fiber.Ctx) (string, error) {
	userID := c.Locals("userID")
	idStr, ok := userID.(string)
	if !ok || idStr == "" {
		return "", errors.New("Invalid user ID")
	}
	return idStr, nil
}

func MarketOrder(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	var body struct {
		Symbol   string `json:"symbol"`
		Side     string `json:"side"`
		Quantity string `json:"quantity"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	order, err := services.PlaceMarketOrder(body.Symbol, body.Side, body.Quantity, userID)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", order)
}

func MarketOrderTPSL(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	var body struct {
		Symbol          string `json:"symbol"`
		Side            string `json:"side"`
		Quantity        string `json:"quantity"`
		TakeProfitPrice string `json:"takeProfitPrice"`
		StopLossPrice   string `json:"stopLossPrice"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	order, oco, err := services.PlaceMarketOrderWithTPSL(body.Symbol, body.Side, body.Quantity, body.TakeProfitPrice, body.StopLossPrice, userID)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", fiber.Map{
		"order": order,
		"oco":   oco,
	})
}

func LimitOrder(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	var body struct {
		Symbol   string `json:"symbol"`
		Side     string `json:"side"`
		Quantity string `json:"quantity"`
		Price    string `json:"price"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	order, err := services.PlaceLimitOrder(body.Symbol, body.Side, body.Quantity, body.Price, userID)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", order)
}

func LimitOrderTPSL(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	var body struct {
		Symbol          string `json:"symbol"`
		Side            string `json:"side"`
		Quantity        string `json:"quantity"`
		Price           string `json:"price"`
		TakeProfitPrice string `json:"takeProfitPrice"`
		StopLossPrice   string `json:"stopLossPrice"`
	}
	if err := c.BodyParser(&body); err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	order, err := services.PlaceLimitOrderWithTPSL(body.Symbol, body.Side, body.Quantity, body.Price, body.TakeProfitPrice, body.StopLossPrice, userID)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", fiber.Map{
		"order": order,
		"tp":    body.TakeProfitPrice,
		"sl":    body.StopLossPrice,
	})
}
