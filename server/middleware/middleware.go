package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"exdex/server/constant"
	"exdex/server/jwt"
)

// JWTMiddleware validates the token and attaches claims to the context
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":   constant.UNAUTHORIZED,
				"status": false,
				"error":  "Missing or invalid Authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		valid, err := jwt.ValidateToken(tokenString)
		if err != nil || !valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":   constant.UNAUTHORIZED,
				"status": false,
				"error":  "Invalid or expired token",
			})
		}

		id, email, role, err := jwt.ExtractClaims(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":   constant.UNAUTHORIZED,
				"status": false,
				"error":  "Failed to extract token claims",
			})
		}

		fmt.Println(">>>")
		fmt.Println(">>>", id)
		fmt.Println(">>>")

		c.Locals("userID", id)
		c.Locals("email", email)
		c.Locals("role", role)

		return c.Next()
	}
}

func RoleMiddleware(expectedRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != expectedRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":   constant.UNAUTHORIZED,
				"status": false,
				"error":  "Access denied: insufficient permissions",
			})
		}
		return c.Next()
	}
}
