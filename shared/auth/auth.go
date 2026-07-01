package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/verasthiago/verancial/shared/models"
)

// UserClaims holds only the minimal, non-sensitive user information that is
// safe to embed in a JWT. JWTs are base64-encoded, NOT encrypted, so secrets
// such as the password hash or bank/financial app credentials must never be
// placed here.
type UserClaims struct {
	ID      string `json:"id"`
	IsAdmin bool   `json:"isadmin"`
}

type JWTClaim struct {
	User *UserClaims
	jwt.StandardClaims
}

func userClaimsFromUser(user *models.User) *UserClaims {
	if user == nil {
		return nil
	}
	return &UserClaims{
		ID:      user.ID,
		IsAdmin: user.IsAdmin,
	}
}

func GenerateJWT(user *models.User, jwtKey string, expirationTime time.Time) (string, error) {
	claims := &JWTClaim{
		User: userClaimsFromUser(user),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))

	return tokenString, err
}

func ValidateToken(signedToken, jwtKey string) error {
	claims, err := GetJWTClaimFromToken(signedToken, jwtKey)
	if err != nil {
		return err
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return errors.New("token expired")
	}

	return nil
}

func GetJWTClaimFromToken(signedToken, jwtKey string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			// Reject tokens signed with an unexpected algorithm (e.g. "none")
			// to prevent algorithm-confusion / signature-bypass attacks.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}

	if claims.User == nil {
		return nil, errors.New("token missing user claims")
	}

	return claims, nil
}
