package router

import (
	"github.com/gofiber/fiber/v2"

	"exdex/internal/src/handler"
)

func (i *impel) orderRouter(r fiber.Router) {
	r.Post("/market", handler.OrderHandler)
	r.Get("/price", handler.GetOrderPriceHandler)
	r.Post("/cancel-oco", handler.GetOrderPriceHandler)
	r.Get("/history", handler.GetOrderHistory)
}
