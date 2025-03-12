package service

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/crypto"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/mutex"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

type UserService struct {
	baseComponent *base.Component
	userDao       *dao.UserDao
	keyLock       mutex.KeyMutex
}

func NewUserService(baseComponent *base.Component, userDao *dao.UserDao, keyLock mutex.KeyMutex) *UserService {
	return &UserService{
		baseComponent: baseComponent,
		userDao:       userDao,
		keyLock:       keyLock,
	}
}

func (s *UserService) Nonce(ctx *reqctx.ReqCtx, req *entity.UserNonceReq) (*entity.UserNonceRes, error) {
	unlock, err := s.keyLock.Lock(mutex.GenerateKey("user", "nonce", req.Addr))
	if err != nil {
		return nil, err
	}
	defer unlock()
	now := time.Now()
	user, err := s.userDao.QueryByAddr(ctx, req.Addr)
	if err != nil && err != errcode.ErrUserNotExist {
		return nil, err
	}
	if user != nil {
		auth, ok := user.Auths[model.UserAuthTypeWallet]
		if !ok {
			nonce := s.baseComponent.UUIDGenerator.Generate().Base58()
			auth = &model.UserAuth{
				Type:  model.UserAuthTypeWallet,
				Token: nonce,
				UID:   req.Addr,
			}
			user.Auths = map[model.UserAuthType]*model.UserAuth{
				model.UserAuthTypeWallet: auth,
			}
		}
		user.LastLoginTime = model.JSONTime(now)
		err = s.userDao.Update(ctx, user)
		if err != nil {
			return nil, err
		}
		return &entity.UserNonceRes{
			Nonce:   auth.Token,
			IssueAt: now.Unix(),
		}, nil
	}

	role := model.UserRoleMember
	if s.baseComponent.IsDevVersion() {
		role = model.UserRoleManager
	}
	nonce := s.baseComponent.UUIDGenerator.Generate().Base58()
	auth := &model.UserAuth{
		Type:  model.UserAuthTypeWallet,
		Token: nonce,
		UID:   req.Addr,
	}
	user = &model.User{
		Addr: req.Addr,
		Auths: map[model.UserAuthType]*model.UserAuth{
			model.UserAuthTypeWallet: auth,
		},
		NickName:      fmt.Sprintf("user-%s", req.Addr[:10]),
		Role:          role,
		Status:        model.UserStatusWaitActive,
		LastLoginTime: model.JSONTime(now),
	}
	err = s.userDao.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return &entity.UserNonceRes{
		Nonce:   nonce,
		IssueAt: now.Unix(),
	}, nil
}

func (s *UserService) Login(ctx *reqctx.ReqCtx, req *entity.UserLoginReq) (*entity.UserLoginRes, error) {
	var user *model.User
	var err error
	switch req.Type {
	case model.UserAuthTypeWallet:
		user, err = s.loginByWallet(ctx, req.UserLoginByWallet)
	default:
		return nil, errors.New("unsupported login auth type")
	}
	if err != nil {
		return nil, err
	}
	return &entity.UserLoginRes{
		UserID:     user.ID.Hex(),
		UserRole:   user.Role,
		UserStatus: user.Status,
	}, nil
}

func (s *UserService) loginByWallet(ctx *reqctx.ReqCtx, req entity.UserLoginByWallet) (*model.User, error) {
	user, err := s.userDao.QueryByAddr(ctx, req.Addr)
	if err != nil && err != errcode.ErrUserNotExist {
		return nil, err
	}
	if err == errcode.ErrUserNotExist {
		return nil, errors.New("user auth nonce not match")
	}

	auth, ok := user.Auths[model.UserAuthTypeWallet]
	if !ok {
		return nil, errors.New("internal error: cannot find user wallet auth nonce")
	}

	if auth.Token != req.Nonce {
		return nil, errors.New("user auth nonce not match")
	}

	// msg := fmt.Sprintf(config.WalletSignatureMessageTemplate, auth.Token)
	msg := req.Message
	err = crypto.VerifyETHSignature(req.Addr, msg, req.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify signature, the nonce may have expired")
	}

	if user.Status == model.UserStatusWaitActive {
		user.Status = model.UserStatusNormal
	}
	user.LastLoginTime = model.JSONTime(time.Now())
	// update nonce
	auth.Token = s.baseComponent.UUIDGenerator.Generate().Base58()
	if err := s.userDao.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DevLogin(ctx *reqctx.ReqCtx, req *entity.UserDevLoginReq) (*entity.UserLoginRes, error) {
	unlock, err := s.keyLock.Lock(mutex.GenerateKey("user", "nonce", req.Addr))
	if err != nil {
		return nil, err
	}
	defer unlock()

	user, err := s.userDao.QueryByAddr(ctx, req.Addr)
	if err != nil && err != errcode.ErrUserNotExist {
		return nil, err
	}
	if user != nil {
		return &entity.UserLoginRes{
			UserID:     user.ID.Hex(),
			UserRole:   user.Role,
			UserStatus: user.Status,
		}, nil
	}

	user = &model.User{
		Addr:     req.Addr,
		NickName: fmt.Sprintf("user-%d", ctx.RequestID),
		Role:     model.UserRoleManager,
		Status:   model.UserStatusNormal,
	}
	if user.Addr != "" && user.Addr == s.baseComponent.Config.App.AdminAddr {
		user.Role = model.UserRoleAdmin
	}
	user.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return nil, err
	}
	user.LastLoginTime = user.CreateTime
	err = s.userDao.Insert(ctx, user)
	if err != nil {
		return nil, err
	}
	return &entity.UserLoginRes{
		UserID:     user.ID.Hex(),
		UserRole:   user.Role,
		UserStatus: user.Status,
	}, nil
}

func (s *UserService) Info(ctx *reqctx.ReqCtx, req *entity.UserInfoReq) (*entity.UserInfoRes, error) {
	user, err := s.userDao.QueryByID(ctx, ctx.Caller)
	if err != nil {
		return nil, err
	}
	return &entity.UserInfoRes{User: user}, nil
}

func (s *UserService) Update(ctx *reqctx.ReqCtx, req *entity.UserUpdateReq) (*entity.UserUpdateRes, error) {
	user, err := s.userDao.QueryByID(ctx, ctx.Caller)
	if err != nil {
		return nil, err
	}
	needUpdate := false
	if req.NewNickName != "" && user.NickName != req.NewNickName {
		user.NickName = req.NewNickName
		needUpdate = true
	}
	if req.NewAvatarURL != "" && user.AvatarURL != req.NewAvatarURL {
		user.AvatarURL = req.NewAvatarURL
		needUpdate = true
	}
	if needUpdate {
		if err := s.userDao.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return &entity.UserUpdateRes{}, nil
}
