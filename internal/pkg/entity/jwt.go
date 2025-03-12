package entity

import (
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/pkg/auth/jwt"
)

type CustomClaims struct {
	jwt.BaseClaims
	CallerRole   model.UserRole   `json:"caller_role"`
	CallerStatus model.UserStatus `json:"caller_status"`
}
