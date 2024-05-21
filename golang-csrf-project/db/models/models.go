package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pirateunclejack/go-practice/golang-csrf-project/randomstrings"
)

type User struct {
    Username, PasswordHash, Role string
}

type TokenClaims struct {
    RegisteredClaims jwt.RegisteredClaims
    Role string `json:"role"`
    Csrf string `json:"csrf"`
}

func (t TokenClaims) Valid() error {
    return nil
}

const RefreshTokenValidTime = time.Hour * 72
const AuthTokenValidTime = time.Minute * 15

func GenerateCSRFSecret() (string, error) {
    return randomstrings.GenerateRandomString(32)
}
