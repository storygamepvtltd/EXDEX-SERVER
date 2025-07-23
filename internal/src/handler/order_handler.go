package handler

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	models "exdex/internal/src/model"
	"exdex/internal/src/services"
	"exdex/server/constant"
	response "exdex/server/responses"
)

func OrderHandler(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	idStr, ok := userID.(string)
	if !ok {
		return response.ErrorMessage(c, constant.BADREQUEST, errors.New("Invalid user ID"))
	}

	var req models.OrderRequest
	if err := c.BodyParser(&req); err != nil {

		return response.ErrorMessage(c, constant.BADREQUEST, err)
	}

	orderServices := services.OrderSerices{}
	result, err := orderServices.PlaceOrder(req, idStr)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", result)
}
func GetOrderPriceHandler(c *fiber.Ctx) error {
	symbol := c.Query("symbol")
	if symbol == "" {
		return response.ErrorMessage(c, constant.BADREQUEST, errors.New("symbol is required"))
	}
	orderServices := services.OrderSerices{}
	p1, err := orderServices.GetSymbolPrice(symbol)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	result := map[string]interface{}{
		"price": p1,
	}

	return response.SuccessResponse(c, "Order sent successfully", result)
}

func GetOrderHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	idStr, ok := userID.(string)
	if !ok {
		return response.ErrorMessage(c, constant.BADREQUEST, errors.New("Invalid user ID"))
	}

	// 🔽 Read limit and offset from query (default to 10 & 0 if not provided)
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	orderServices := services.OrderSerices{}
	data, err := orderServices.GetHistory(idStr, limit, offset)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "", data)
}

func CancelOCOOrderHandler(c *fiber.Ctx) error {
	orderListIdStr := c.Query("orderListId")
	symbol := c.Query("symbol")

	orderListId, err := strconv.ParseInt(orderListIdStr, 10, 64)
	if err != nil {
		return response.ErrorMessage(c, constant.BADREQUEST, errors.New(""))
	}

	orderServices := services.OrderSerices{}
	res, err := orderServices.CancelOCOOrder(orderListId, symbol)
	if err != nil {
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}

	return response.SuccessResponse(c, "Order sent successfully", res)
}

func WebSocketHandler(c *websocket.Conn) {
	defer c.Close()
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		log.Printf("Received: %s", msg)

		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("WebSocket write error:", err)
			break
		}
	}
}
