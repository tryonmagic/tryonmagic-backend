package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyToken(tokenStr, secret, tokenType string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	jwtType, ok := claims["type"].(string)
	if !ok || jwtType != tokenType {
		return "", errors.New("invalid token type")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid token sub")
	}

	return userID, nil
}

func DecodeJWTToken(tokenStr string, secret string) (jwt.MapClaims, error) {
	if secret == "" {
		return parseJWTToken(tokenStr)
	}
	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func parseJWTToken(tokenStr string) (jwt.MapClaims, error) {
	// Parse the token without verifying the signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, errors.New("failed to decode token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func VerifyShopifyToken() {

}
