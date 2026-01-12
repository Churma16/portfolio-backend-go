package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SecretKey adalah Rahasia dapur (Harusnya di .env, tapi buat belajar kita taruh sini dulu atau ambil dr param)
var SecretKey = []byte("rahasia_super_negara_api")

// ErrTokenInvalid adalah error untuk token yang tidak valid
var ErrTokenInvalid = errors.New("token tidak valid")

// CreateToken membuat token JWT yang berlaku selama durasi tertentu
func CreateToken(userID int64, duration time.Duration) (string, error) {
	// 1. Tentukan isi token (Claims)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(), // Kapan kadaluarsa
		"iat":     time.Now().Unix(),               // Kapan dibuat
	}

	// 2. Buat token dengan algoritma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Tanda tangani token dengan Secret Key
	return token.SignedString(SecretKey)
}

// VerifyToken memeriksa apakah token valid dan mengembalikan User ID
func VerifyToken(tokenString string) (int64, error) {
	// 1. Parse token dengan Secret Key yang sama
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pastikan algoritma-nya benar (HMAC)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return SecretKey, nil
	})

	if err != nil {
		return 0, err
	}

	// 2. Ambil data (Claims) dari dalam token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Ambil user_id dan ubah jadi int64
		userID := int64(claims["user_id"].(float64))
		return userID, nil
	}

	return 0, ErrTokenInvalid
}
