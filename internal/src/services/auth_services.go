package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	models "exdex/internal/src/model"
	"exdex/internal/src/repository"
	"exdex/server/jwt"
)

type AuthServices struct{}

type ExdexTokenRequest struct {
	Token string `json:"token"`
}

type ExdexUserResponse struct {
	ID            int    `json:"id"`
	Email         string `json:"email"`
	AccountNumber string `json:"accountNumber"`
	FullName      string `json:"fullName"`
	Role          string `json:"role"`
}

type ExdexResponse struct {
	Code   int               `json:"code"`
	Status string            `json:"status"`
	Data   ExdexUserResponse `json:"data"`
	Error  string            `json:"error,omitempty"`
}

func (a AuthServices) ExdexAuth(exdexToken string) (string, error) {
	url := fmt.Sprintf("https://api.exdex.com/api/exdex/check?token=%s", exdexToken)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// fmt.Println("🔍 Raw Response Body:", string(body))

	var apiResp ExdexResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if apiResp.Code != 200 {
		return "", fmt.Errorf("unexpected status code: %d, error: %s", apiResp.Code, apiResp.Error)
	}

	var user models.User
	user.ExdexUserID = uint(apiResp.Data.ID)
	user.Email = apiResp.Data.Email
	user.AccountNumber = apiResp.Data.AccountNumber
	user.FullName = apiResp.Data.FullName
	user.CreatedAt = time.Now()
	user.Status = true
	user.Role = apiResp.Data.Role

	err = repository.IRepo.Create("users", &user)
	if err != nil {
		if strings.Contains(err.Error(), "E11000 duplicate key error") {
			fmt.Errorf("duplicate: user with this email already exists")
		} else {
			return "", err
		}
	}

	token, err := jwt.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}
