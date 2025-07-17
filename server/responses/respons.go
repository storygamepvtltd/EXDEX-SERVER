package response

import (
	"github.com/gofiber/fiber/v2"
)

type Success struct {
	Code    int         `json:"code"`
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Connections struct {
	Code     int    `json:"code"`
	Status   bool   `json:"status"`
	Message  string `json:"message"`
	UserName string `json:"user_name"`
}

type Error struct {
	Code   int    `json:"code"`
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

type Warning struct {
	Code    int    `json:"code"`
	Status  bool   `json:"status"`
	Resp    string `json:"response"`
	Message string `json:"message"`
}

type TwitterTokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	result := Success{
		Code:    200,
		Status:  true,
		Message: message,
		Data:    data,
	}
	return JsonResponse(c, 200, result)
}

func ErrorMessage(c *fiber.Ctx, code int, e error) error {
	errMsg := Error{
		Status: false,
		Code:   code,
		Error:  e.Error(),
	}
	return JsonResponse(c, code, errMsg)
}

func WarningMessage(c *fiber.Ctx, resp, msg string) error {
	warning := Warning{
		Code:    409,
		Status:  true,
		Resp:    resp,
		Message: msg,
	}
	return JsonResponse(c, 409, warning)
}

func JsonResponse(c *fiber.Ctx, code int, response interface{}) error {
	if response != nil {
		return c.Status(code).JSON(response)
	}
	return c.SendStatus(code)
}

func ConnectionSuccessResponse(c *fiber.Ctx, message string, username string) error {
	result := Connections{
		Code:     200,
		Status:   true,
		Message:  message,
		UserName: username,
	}
	return JsonResponse(c, 200, result)
}
