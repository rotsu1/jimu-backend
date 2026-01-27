package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTokenPair creates a short-lived access token and a long-lived refresh token.
func GenerateTokenPair(userId string, secret string) (string, string, error) {
	// 1. Access Token (15 Minutes)
	atClaims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		atClaims,
	).SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (1 year)
	rtClaims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour * 24 * 365).Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		rtClaims,
	).SignedString([]byte(secret))

	return accessToken, refreshToken, err
}

func VerifyToken(tokenString string, secret string) (string, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is expected (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	// Get the claims (contents)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Get the userID (sub) that was set when the JWT was generated
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("user_id (sub) not found in token")
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
