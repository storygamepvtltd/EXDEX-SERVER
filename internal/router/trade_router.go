package router

import (
	"github.com/gofiber/fiber/v2"

	"exdex/internal/src/handler"
)

func (i *impel) tradeRouter(r fiber.Router, app *fiber.App) {
	// r.Get("/market/tickers", handler.RegisterWebSocket(app))
	handler.RegisterWebSocket(app)
}
