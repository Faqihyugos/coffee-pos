package middleware

import (
	"fmt"
	"strings"

	"github.com/faqihyugos/coffee-pos/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Token tidak ditemukan")
			c.Abort()
			return
		}

		// 2. Cek prefix "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "Format token tidak valid")
			c.Abort()
			return
		}

		// 3. Potong prefix "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 4. Parse dan validasi token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validasi signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("signing method tidak valid: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Token tidak valid atau sudah expired")
			c.Abort()
			return
		}

		// 5. Ekstrak claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		// 6. Ambil user_id dan role
		userID, okID := claims["user_id"].(string)
		role, okRole := claims["role"].(string)

		if !okID || !okRole {
			response.Unauthorized(c, "Payload token tidak valid")
			c.Abort()
			return
		}

		// 7. Simpan ke context
		c.Set("user_id", userID)
		c.Set("role", role)

		// 8. Lanjut ke handler berikutnya
		c.Next()
	}
}

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			response.Unauthorized(c, "Token tidak valid")
			c.Abort()
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if r == role {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "Akses ditolak")
			c.Abort()
			return
		}

		c.Next()
	}
}
