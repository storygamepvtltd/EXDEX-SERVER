package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"exdex/internal/src/handler"
)

func (i *impel) orderRouter(r fiber.Router) {
	r.Post("/market", handler.OrderHandler)
	r.Get("/price", handler.GetOrderPriceHandler)
	r.Post("/cancel-oco", handler.GetOrderPriceHandler)
	r.Get("/history", handler.GetOrderAllHistory)

	r.Post("/market-order", handler.MarketOrder)
	r.Post("/market-order-tpsl", handler.MarketOrderTPSL)
	r.Post("/limit-order", handler.LimitOrder)
	r.Post("/limit-order-tpsl", handler.LimitOrderTPSL)

	// inside Start()
	r.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

}
