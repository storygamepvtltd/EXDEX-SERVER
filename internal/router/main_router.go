package router

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"

	"exdex/server/middleware"
)

type Ctx = *fiber.Ctx

type Router interface {
	Start()
}

type impel struct {
	fiver *fiber.App
}

func (i *impel) Start() {
	app := i.fiver
	app.Get("/", func(c Ctx) error {
		return c.SendString("Hello, World!")
	})
	// Example group
	v1 := app.Group("/v1")
	{
		api := v1.Group("api")
		{
			i.authRouter(api)
		}

		auth := v1.Group("auth", middleware.JWTMiddleware())
		{
			trade := auth.Group("tarde")
			{
				i.tradeRouter(trade, app)
			}
			order := auth.Group("order")
			{
				i.orderRouter(order)
			}
		}

	}

	// addr := fmt.Sprintf(":%s", viper.GetString("server.port"))

	// Start server in goroutine
	// go func() {
	// 	fmt.Printf("🚀 Server starting on port %s\n", viper.GetString("server.port"))
	// 	if err := app.Listen(addr); err != nil {
	// 		log.Printf("❌ Server error: %v\n", err)
	// 	}
	// }()

	// ✅ Print all routes
	PrintEndpoints(app)
	log.Fatal(i.fiver.Listen(fmt.Sprintf(":%s", viper.GetString("server.port"))))
	// Wait for shutdown signal and call graceful shutdown
	// waitForShutdown(app)
}

func PrintEndpoints(app *fiber.App) {
	fmt.Println("📍 Registered Endpoints:")
	for _, route := range app.Stack() {
		for _, r := range route {
			fmt.Printf("➡️  %-6s %s\n", r.Method, r.Path)
		}
	}
}

// Separate function for graceful shutdown

func NewRouter() Router {
	return &impel{
		fiver: fiber.New(),
	}
}

// func waitForShutdown(app *fiber.App) {
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
// 	<-quit

// 	fmt.Println("🛑 Shutdown signal received")

// 	// Context with timeout for cleanup
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := app.Shutdown(); err != nil {
// 		log.Printf("❌ Error during shutdown: %v\n", err)
// 	}

// 	select {
// 	case <-ctx.Done():
// 		fmt.Println("✅ Server gracefully stopped")
// 	}
// }
