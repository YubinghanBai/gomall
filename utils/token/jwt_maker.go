package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid secret key size: must be at least %d bytes", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(userID int64, username string, role string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(userID, username, role, duration)
	if err != nil {
		return "", nil, err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         payload.ID,
		"user_id":    userID,
		"username":   username,
		"role":       role,
		"issued_at":  time.Now().Unix(),
		"expired_at": time.Now().Add(duration).Unix(),
	})

	tokenString, err := jwtToken.SignedString([]byte(maker.secretKey))
	return tokenString, payload, err
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrInvalidToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
