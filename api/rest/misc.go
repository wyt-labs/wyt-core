package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) chainList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChainListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)

	res, err := s.MiscService.ChainList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) trackList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TrackListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)
	ctx.AddCustomLogField("hot", req.IsHot)

	res, err := s.MiscService.TrackList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) tagList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TagListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)
	ctx.AddCustomLogField("hot", req.IsHot)

	res, err := s.MiscService.TagList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) tagAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TagAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("name", req.Name)
	ctx.AddCustomLogField("score", req.Score)

	res, err := s.MiscService.TagAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) teamImpressionList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TeamImpressionListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)

	res, err := s.MiscService.TeamImpressionList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) investorList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.InvestorListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)

	res, err := s.MiscService.InvestorList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) investorAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.InvestorAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("name", req.Name)
	ctx.AddCustomLogField("avatar", req.AvatarURL)
	ctx.AddCustomLogField("subject", req.Subject)
	ctx.AddCustomLogField("type", req.Type)

	res, err := s.MiscService.InvestorAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) investorUpdate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.InvestorUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("name", req.Name)
	ctx.AddCustomLogField("avatar", req.AvatarURL)
	ctx.AddCustomLogField("subject", req.Subject)
	ctx.AddCustomLogField("type", req.Type)

	res, err := s.MiscService.InvestorUpdate(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
