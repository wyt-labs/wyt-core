package rest

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/auth/jwt"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) userNonce(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserNonceReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if req.Addr == "" {
		return nil, errcode.ErrRequestParameter.Wrap("addr cannot be empty")
	}
	if !common.IsHexAddress(req.Addr) {
		return nil, errcode.ErrRequestParameter.Wrap("addr is invalid")
	}

	ctx.AddCustomLogField("addr", req.Addr)

	res, err := s.UserService.Nonce(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) userSignin(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserLoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	switch req.Type {
	case model.UserAuthTypeWallet:
		if req.Addr == "" {
			return nil, errcode.ErrRequestParameter.Wrap("addr cannot be empty")
		}
		if req.Signature == "" {
			return nil, errcode.ErrRequestParameter.Wrap("signature cannot be empty")
		}
		if !common.IsHexAddress(req.Addr) {
			return nil, errcode.ErrRequestParameter.Wrap("addr is invalid")
		}
		ctx.AddCustomLogField("addr", req.Addr)
	}

	res, err := s.UserService.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("caller", res.UserID)

	res.Token, res.ExpiredDate, err = jwt.GenerateWithHMACKey(s.baseComponent.Config.HTTP.JWTTokenHMACKey, s.baseComponent.Config.HTTP.JWTTokenValidDuration.ToDuration(), res.UserID, &entity.CustomClaims{
		CallerRole:   res.UserRole,
		CallerStatus: res.UserStatus,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) userLogin(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserLoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	switch req.Type {
	case model.UserAuthTypeWallet:
		if req.Addr == "" {
			return nil, errcode.ErrRequestParameter.Wrap("addr cannot be empty")
		}
		if req.Signature == "" {
			return nil, errcode.ErrRequestParameter.Wrap("signature cannot be empty")
		}
		if !common.IsHexAddress(req.Addr) {
			return nil, errcode.ErrRequestParameter.Wrap("addr is invalid")
		}
		ctx.AddCustomLogField("addr", req.Addr)
	}

	res, err := s.UserService.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("caller", res.UserID)

	res.Token, res.ExpiredDate, err = jwt.GenerateWithHMACKey(s.baseComponent.Config.HTTP.JWTTokenHMACKey, s.baseComponent.Config.HTTP.JWTTokenValidDuration.ToDuration(), res.UserID, &entity.CustomClaims{
		CallerRole:   res.UserRole,
		CallerStatus: res.UserStatus,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) userDevLogin(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserDevLoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	if req.Addr == "" {
		return nil, errcode.ErrRequestParameter.Wrap("addr cannot be empty")
	}
	if !common.IsHexAddress(req.Addr) {
		return nil, errcode.ErrRequestParameter.Wrap("addr is invalid")
	}

	res, err := s.UserService.DevLogin(ctx, req)
	if err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("caller", res.UserID)

	res.Token, res.ExpiredDate, err = jwt.GenerateWithHMACKey(s.baseComponent.Config.HTTP.JWTTokenHMACKey, s.baseComponent.Config.HTTP.JWTTokenValidDuration.ToDuration(), res.UserID, &entity.CustomClaims{
		CallerRole:   res.UserRole,
		CallerStatus: res.UserStatus,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) userInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	res, err := s.UserService.Info(ctx, &entity.UserInfoReq{})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) userUpdate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.UserUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.NewNickName == "" {
		return nil, errcode.ErrRequestParameter.Wrap("new_nick_name cannot be empty")
	}
	if req.NewAvatarURL == "" {
		return nil, errcode.ErrRequestParameter.Wrap("new_avatar_url cannot be empty")
	}

	res, err := s.UserService.Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) userRefreshToken(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	token, expiredDate, err := jwt.GenerateWithHMACKey(s.baseComponent.Config.HTTP.JWTTokenHMACKey, s.baseComponent.Config.HTTP.JWTTokenValidDuration.ToDuration(), ctx.Caller, &entity.CustomClaims{
		CallerRole:   ctx.CallerRole,
		CallerStatus: ctx.CallerStatus,
	})
	if err != nil {
		return nil, err
	}
	return &entity.UserLoginRes{
		Token:       token,
		ExpiredDate: expiredDate,
	}, nil
}
