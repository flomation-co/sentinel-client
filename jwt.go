package sentinel

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

func Sentinel(identityService string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if headerParts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userID, err := GetUser(identityService, headerParts[1])
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if userID == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("flomation-user-id", *userID)

		c.Next()
	}
}

func GetUser(identityService string, jwt string) (*string, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/user", identityService), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New("invalid response code")
	}

	if res.Body == nil {
		return nil, errors.New("invalid body")
	}

	defer func() {
		_ = res.Body.Close()
	}()

	type UserResponse struct {
		UserID string `json:"user_id"`
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var response UserResponse
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}

	return &response.UserID, nil
}
