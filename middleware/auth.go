package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the JWT payload structure.
type Claims struct {
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

// SignToken handler – generates a JWT from PER_CITIZEN_ID (mirrors auth_sign.js sign_mid).
// POST /api/service/frontend/sign
func SignToken(c *gin.Context) {
	var body struct {
		PERCitizenID string `json:"PER_CITIZEN_ID"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.PERCitizenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "กรุณาระบุ PER_CITIZEN_ID",
		})
		return
	}

	secretKey := os.Getenv("secrete_id")
	if secretKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "secret key not configured"})
		return
	}

	claims := &Claims{
		ClientID: body.PERCitizenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secretKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "ไม่สามารถสร้าง Token ได้"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "token": tokenStr})
}

// VerifyToken middleware – validates Bearer JWT and sets decoded claims in context.
// Mirrors auth_sign.js verify_mid.
func VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized Access. Missing Bearer Token.",
		})
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized Access. Token is empty.",
		})
		return
	}

	secretKey := os.Getenv("secrete_id")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authorized Expire.",
		})
		return
	}

	// Store decoded info for downstream handlers
	c.Set("decoded_client_id", claims.ClientID)
	c.Next()
}

// VerifyTokenEndpoint – standalone POST /verify handler (returns success/fail JSON).
func VerifyTokenEndpoint(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized Access. Missing Bearer Token."})
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secretKey := os.Getenv("secrete_id")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorized Expire."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "decoded": claims.ClientID})
}
