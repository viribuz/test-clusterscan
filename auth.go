package main

import "fmt"
import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

// GenerateJWT generates a new JWT token for an agent
func GenerateJWT(agentID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"agentID": agentID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your_secret_key")) // Replace with your secret key
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// AuthenticateAgent checks the credentials and returns a JWT token if they are valid
func AuthenticateAgent(username, password string) (string, error) {
	agentID, valid := VerifyAgentCredentials(username, password)
	if !valid {
		return "", fmt.Errorf("invalid credentials")
	}
	return GenerateJWT(agentID)
}
// RenewJWT checks the current token and renews it if it's expired
func RenewJWT(tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your_secret_key"), nil
	})

	if err != nil {
		return "", err
	}

	// Check if token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check if token is expired
		if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
			// Token is expired, issue a new one
			agentID, ok := claims["agentID"].(string)
			if !ok {
				return "", fmt.Errorf("invalid token claims")
			}
			return GenerateJWT(agentID)
		}
		return "", fmt.Errorf("token is not expired")
	}

	return "", fmt.Errorf("invalid token")
}
