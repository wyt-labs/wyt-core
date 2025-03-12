package service

import (
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *UserService) AdminSetRole(ctx *reqctx.ReqCtx, req *entity.UserSetRoleReq) (*entity.UserSetRoleRes, error) {
	if req.Role == model.UserRoleAdmin {
		return nil, errcode.ErrAccountPermission
	}

	user, err := s.userDao.QueryByID(ctx, req.TargetUserID)
	if err != nil {
		return nil, err
	}
	needUpdate := false
	if user.Role != req.Role {
		user.Role = req.Role
		needUpdate = true
	}
	if needUpdate {
		if err := s.userDao.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return &entity.UserSetRoleRes{}, nil
}

func (s *UserService) AdminSetIsLock(ctx *reqctx.ReqCtx, req *entity.UserSetIsLockReq) (*entity.UserSetIsLockRes, error) {
	user, err := s.userDao.QueryByID(ctx, req.TargetUserID)
	if err != nil {
		return nil, err
	}
	needUpdate := false
	if req.IsLock && user.Status != model.UserStatusLocked {
		user.Status = model.UserStatusLocked
		needUpdate = true
	} else if !req.IsLock && user.Status == model.UserStatusLocked {
		user.Status = model.UserStatusNormal
		needUpdate = true
	}
	if needUpdate {
		if err := s.userDao.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return &entity.UserSetIsLockRes{}, nil
}
