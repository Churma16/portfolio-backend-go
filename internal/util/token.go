package util

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HAPUS variabel global SecretKey!
// var SecretKey = []byte("rahasia_super_negara_api") <-- BUANG INI

var (
	ErrTokenInvalid = errors.New("token tidak valid")
	ErrInvalidKey   = errors.New("TOKEN_SYMMETRIC_KEY must be exactly 32 characters and cannot be empty or default fallback")
)

const InsecureDefaultKey = "rahasia_super_negara_api"

// Update: Tambahkan parameter 'secretKey string'
func CreateToken(userID int64, duration time.Duration, secretKey string) (string, error) {
	if len(secretKey) != 32 || secretKey == InsecureDefaultKey {
		return "", ErrInvalidKey
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Gunakan secretKey dari parameter (convert ke []byte)
	return token.SignedString([]byte(secretKey))
}

// Update: Tambahkan parameter 'secretKey string'
func VerifyToken(tokenString string, secretKey string) (int64, error) {
	if len(secretKey) != 32 || secretKey == InsecureDefaultKey {
		return 0, ErrInvalidKey
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		// Gunakan secretKey dari parameter
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Konversi aman float64 ke int64
		if val, ok := claims["user_id"].(float64); ok {
			return int64(val), nil
		}
		return 0, fmt.Errorf("invalid user_id type")
	}

	return 0, ErrTokenInvalid
}
