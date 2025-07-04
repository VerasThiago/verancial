package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/verasthiago/verancial/shared/models"
)

type JWTClaim struct {
	User *models.User
	jwt.StandardClaims
}

func GenerateJWT(user *models.User, jwtKey string, expirationTime time.Time) (string, error) {
	claims := &JWTClaim{
		User: user,
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

	return claims, nil
}
