package jwt

import (
	"time"
)

type Claims struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

type TokenPair struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}