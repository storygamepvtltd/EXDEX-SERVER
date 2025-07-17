package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"exdex/internal/src/services"
	"exdex/server/constant"
	response "exdex/server/responses"
)

func CheckExdexUserInfo(c *fiber.Ctx) error {
	exdexToken := c.Query("token")
	if exdexToken == "" {
		return response.ErrorMessage(c, constant.BADREQUEST, errors.New("token is required"))
	}
	exdexServices := services.AuthServices{}
	user, err := exdexServices.ExdexAuth(exdexToken)
	if err != nil {
		log.Println("Service error:", err)
		return response.ErrorMessage(c, constant.INTERNALSERVERERROR, err)
	}
	return response.SuccessResponse(c, "Successfully fetched user in EXDEX", user)
}
