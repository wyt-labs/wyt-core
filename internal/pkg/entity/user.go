package entity

import (
	"github.com/wyt-labs/wyt-core/internal/core/model"
)

type UserNonceReq struct {
	Addr string `json:"addr" form:"addr"`
}

type UserNonceRes struct {
	Nonce   string `json:"nonce"`
	IssueAt int64  `json:"issue_at"`
}

type UserLoginByWallet struct {
	Addr      string `json:"addr"`
	Signature string `json:"signature"`
	Nonce     string `json:"nonce"`
	Message   string `json:"message"`
}

type UserLoginReq struct {
	Type model.UserAuthType `json:"type"`
	UserLoginByWallet
}

type UserDevLoginReq struct {
	Addr string `json:"addr"`
}

type UserLoginRes struct {
	UserID      string           `json:"-"`
	UserRole    model.UserRole   `json:"-"`
	UserStatus  model.UserStatus `json:"-"`
	Token       string           `json:"token"`
	ExpiredDate int64            `json:"expired_date"`
}

type UserInfoReq struct {
}

type UserInfoRes struct {
	*model.User
}

type UserUpdateReq struct {
	NewNickName  string `json:"new_nick_name" form:"new_avatar_url"`
	NewAvatarURL string `json:"new_avatar_url" form:"new_avatar_url"`
}

type UserUpdateRes struct {
}

type UserSetRoleReq struct {
	TargetUserID string         `json:"target_user_id"`
	Role         model.UserRole `json:"role"`
}

type UserSetRoleRes struct {
}

type UserSetIsLockReq struct {
	TargetUserID string `json:"target_user_id"`
	IsLock       bool   `json:"is_lock"`
}

type UserSetIsLockRes struct {
}
