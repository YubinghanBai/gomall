package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gomall/utils/response"
	"gomall/utils/token"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			response.Error(c, http.StatusUnauthorized, "authorization header is not provided")
			c.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			response.Error(c, http.StatusUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}
		authorizationType := fields[0]
		if authorizationType != authorizationTypeBearer {
			response.Error(c, http.StatusUnauthorized, "invalid authorization type")
			c.Abort()
			return
		}
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			if errors.Is(err, token.ErrExpiredToken) {
				response.Error(c, http.StatusUnauthorized, "token has expired")
			} else {
				response.Error(c, http.StatusUnauthorized, "invalid token")
			}
			c.Abort()
			return
		}
		c.Set(authorizationPayloadKey, payload)
		c.Next()
	}
}

func GetPayload(c *gin.Context) *token.Payload {
	value, exists := c.Get(authorizationPayloadKey)
	if !exists {
		return nil
	}
	payload, ok := value.(*token.Payload)
	if !ok {
		return nil
	}
	return payload
}
