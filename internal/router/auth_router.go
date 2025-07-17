package router

import (
	"github.com/gofiber/fiber/v2"

	"exdex/internal/src/handler"
)

func (i *impel) authRouter(r fiber.Router) {
	r.Post("/exdex/check", handler.CheckExdexUserInfo)
}
