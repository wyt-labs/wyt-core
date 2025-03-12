package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) dataPumpNewTokens(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.NewTokens(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpLaunchTime(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.LaunchTime(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTransactions(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.Transactions(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTopTraders(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TopTraders(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderInfo(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TraderInfo(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderOverview(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TraderOverviewV2(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderProfit(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TraderProfit(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderProfitDistribution(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TraderProfitDistribution(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderTrades(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	if res, err = s.CoreAPI.PumpDataService.TraderTrades(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) dataPumpTraderDetail(ctx *reqctx.ReqCtx, c *gin.Context) (res any, err error) {
	req := &model.CommonPumpDataQuery{}
	if err = c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	// default last 7 days of UTC
	if req.Duration == 0 {
		req.Duration = 7
	}
	if req.Timezone == "" {
		req.Timezone = "CST"
	}
	if res, err = s.CoreAPI.PumpDataService.TraderDetail(ctx.Ctx, req); err != nil {
		return nil, err
	}
	return res, nil
}
