package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) adminChainAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChainAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("chain", req.Name)

	res, err := s.MiscService.AdminChainAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminTrackAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TrackAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("track", req.Name)

	res, err := s.MiscService.AdminTrackAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminTeamImpressionAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.TeamImpressionAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("team_impression", req.Name)

	res, err := s.MiscService.AdminTeamImpressionAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
