package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/verasthiago/verancial/api/pkg/builder"
	"github.com/verasthiago/verancial/shared/auth"
	shared "github.com/verasthiago/verancial/shared/flags"
)

type AuthUserAPI interface {
	Handler() gin.HandlerFunc
}

type AuthUserHandler struct {
	*builder.Flags
	*shared.SharedFlags
}

func (a *AuthUserHandler) InitFromFlags(flags *builder.Flags, sharedFlags *shared.SharedFlags) *AuthUserHandler {
	a.Flags = flags
	a.SharedFlags = sharedFlags

	return a
}

func (a *AuthUserHandler) Handler() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err := auth.ValidateToken(tokenString, a.JwtKey)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		jwtClaim, err := auth.GetJWTClaimFromToken(tokenString, a.JwtKey)
		if err != nil {
			context.JSON(401, gin.H{"error": "invalid token"})
			context.Abort()
			return
		}
		context.Set("user", jwtClaim.User)
		context.Next()
	}
}
