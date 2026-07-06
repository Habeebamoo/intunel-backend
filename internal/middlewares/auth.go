package middlewares

import (
	"net/http"

	"github.com/Habeebamoo/intunel-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("tunnl_token")
		if err != nil || tokenStr == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "missing or invalid token")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}