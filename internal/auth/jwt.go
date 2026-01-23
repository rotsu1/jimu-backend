package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userId string, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
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
