package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) adminUserSetRole(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserSetRoleReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.TargetUserID == "" {
		return nil, errcode.ErrAccountPermission.Wrap("target_user_id cannot be empty")
	}
	if req.TargetUserID == ctx.Caller {
		return nil, errcode.ErrAccountPermission.Wrap("target_user_id cannot be caller self")
	}
	if s.baseComponent.Config.App.AdminAddr == req.TargetUserID {
		return nil, errcode.ErrAccountPermission.Wrap("cannot set admin role")
	}

	ctx.AddCustomLogField("target_user", req.TargetUserID)
	ctx.AddCustomLogField("role", req.Role)

	res, err := s.UserService.AdminSetRole(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminUserSetIsLock(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserSetIsLockReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.TargetUserID == "" {
		return nil, errcode.ErrAccountPermission.Wrap("target_user_id cannot be empty")
	}
	if req.TargetUserID == ctx.Caller {
		return nil, errcode.ErrAccountPermission.Wrap("target_user_id cannot be caller self")
	}
	if s.baseComponent.Config.App.AdminAddr == req.TargetUserID {
		return nil, errcode.ErrAccountPermission.Wrap("cannot set admin role")
	}
	ctx.AddCustomLogField("target_user", req.TargetUserID)
	ctx.AddCustomLogField("is_lock", req.IsLock)

	res, err := s.UserService.AdminSetIsLock(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
