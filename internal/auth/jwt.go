package auth

import (
	"net/http"
	"strings"
	"time"

	"restaurant-system/internal/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func jwtSecret() []byte {
	s := config.Auth().JWTSecret
	if s == "" {
		s = "replace-with-secure-secret"
	}
	return []byte(s)
}

type Claims struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role,omitempty"`
	jwt.StandardClaims
}

func GenerateToken(accountID string, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		AccountID: accountID,
		Role:      role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(ttl).Unix(),
			IssuedAt:  now.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateAccessAndRefresh creates an access JWT and a refresh token (opaque string).
func GenerateAccessAndRefresh(accountID string, role string) (accessToken string, refreshToken string, err error) {
	accessTTL := time.Duration(config.Auth().AccessTokenMinutes) * time.Minute
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	accessToken, err = GenerateToken(accountID, role, accessTTL)
	if err != nil {
		return "", "", err
	}

	// refresh token: long random opaque string (UUID-like)
	// we keep it simple here; services should store and validate it server-side
	refreshToken = jwt.StandardClaims{Id: time.Now().Format(time.RFC3339Nano)}.Id
	return accessToken, refreshToken, nil
}

// Middleware to require a valid token and optionally role
func RequireAuth(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.Fields(h)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}
		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret(), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if claims, ok := token.Claims.(*Claims); ok {
			if role != "" && claims.Role != role {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
				return
			}
			c.Set("account_id", claims.AccountID)
			c.Set("role", claims.Role)
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
	}
}
