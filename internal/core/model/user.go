package model

import (
	"github.com/pkg/errors"
)

type UserStatus = uint32

const (
	UserStatusWaitActive UserStatus = iota
	UserStatusNormal
	UserStatusLocked
	UserStatusCancelled
)

type UserRole uint32

const (
	UserRoleMember UserRole = iota
	UserRoleManager
	UserRoleAdmin
)

func (r UserRole) Validate() error {
	if uint32(r) > uint32(UserRoleAdmin) {
		return errors.New("invalid user role")
	}
	return nil
}

type UserAuthType uint32

const (
	UserAuthTypeWallet UserAuthType = iota
	UserAuthTypeGoogle
)

func (t UserAuthType) Validate() error {
	if uint32(t) > uint32(UserAuthTypeGoogle) {
		return errors.New("invalid user auth type")
	}
	return nil
}

type UserAuth struct {
	Type UserAuthType `json:"type" bson:"type"`

	// UserAuthTypeWallet: address
	UID string `json:"uid" bson:"uid"`

	// UserAuthTypeWallet: nonce
	Token string `json:"token" bson:"token"`

	Extra map[string]string `json:"extra" bson:"extra"`
}

type User struct {
	BaseModel     `bson:"inline"`
	Addr          string                     `json:"addr" bson:"addr"`
	Email         string                     `json:"email" bson:"email"`
	Phone         string                     `json:"phone" bson:"phone"`
	Password      string                     `json:"-" bson:"password"`
	Auths         map[UserAuthType]*UserAuth `json:"-" bson:"auths"`
	NickName      string                     `json:"nick_name" bson:"nick_name"`
	AvatarURL     string                     `json:"avatar_url" bson:"avatar_url"`
	Role          UserRole                   `json:"role" bson:"role"`
	Status        UserStatus                 `json:"status" bson:"status"`
	LastLoginTime JSONTime                   `json:"last_login_time" bson:"last_login_time"`
}
