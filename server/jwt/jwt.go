package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

func GenerateJWT(userID string, email string, role string) (string, error) {
	expiration, err := time.ParseDuration(viper.GetString("jwt.expiration"))
	if err != nil {
		return "", err
	}
	// fmt.Println("lsfkjglsfjg", viper.GetString("jwt.expiration"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userID,
		"email": email,
		"role":  role, // Include role in the JWT token
		"exp":   time.Now().Add(expiration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func ExtractClaims(tokenString string) (string, string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("jwt.secret")), nil // Use correct secret key
	})

	if err != nil {
		return "", "", "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id, okID := claims["id"].(string) // JWT stores numbers as float64
		email, okEmail := claims["email"].(string)
		role, okRole := claims["role"].(string)

		if !okID || !okEmail || !okRole {
			return "", "", "", fmt.Errorf("invalid claim format")
		}

		return id, email, role, nil
	}

	return "", "", "", fmt.Errorf("invalid token claims")
}

func ValidateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(viper.GetString("jwt.secret")), nil
	})

	if err != nil {
		return false, err
	}

	return token.Valid, nil
}

// ValidateRole checks if the JWT token has the expected role
func ValidateRole(tokenString, expectedRole string) (bool, error) {
	_, _, role, err := ExtractClaims(tokenString)
	if err != nil {
		return false, err
	}

	if role != expectedRole {
		return false, fmt.Errorf("unauthorized role: expected %s but got %s", expectedRole, role)
	}

	return true, nil
}
