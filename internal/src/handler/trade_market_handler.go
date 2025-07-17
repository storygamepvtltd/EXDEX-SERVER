package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"exdex/internal/src/services"
)

func RegisterWebSocket(app *fiber.App) {
	app.Get("/ws/market/tickers", websocket.New(func(c *websocket.Conn) {
		defer c.Close()

		tickerChan := make(chan services.StreamTicker)

		// Start Binance stream in a goroutine
		go services.StartBinanceTickerStream(tickerChan)

		for ticker := range tickerChan {
			if err := c.WriteJSON(ticker); err != nil {
				log.Println("WebSocket send error:", err)
				break
			}
		}
	}))
}
