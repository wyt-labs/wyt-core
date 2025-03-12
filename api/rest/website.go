package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) websiteSubscribe(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.WebsiteSubscribeReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.Email == "" {
		return nil, errcode.ErrRequestParameter.Wrap("email cannot be empty")
	}
	ctx.AddCustomLogField("email", req.Email)

	res, err := s.WebsiteService.Subscribe(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) websiteUnsubscribe(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.WebsiteUnsubscribeReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("email", req.Email)

	res, err := s.WebsiteService.Unsubscribe(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
