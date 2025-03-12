package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/wyt-labs/wyt-core/internal/core/component/datapuller/model"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) dexSupportedChain(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetSupportedChainsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetSupportedChains(req.ChainId, false)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) dexSupportedAllTokens(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetSupportedTokensReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetTokens(req.ChainId)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) tokenInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetTokenReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetToken(req.ChainId, req.TokenAddress)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) getQuote(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetQuoteReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetQuotes(req.ChainId, req.Amount, req.FromTokenAddress, req.ToTokenAddress, "")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) approveTransaction(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.ApproveTransactionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.ApproveTransaction(req.ChainId, req.TokenContractAddress, req.ApproveAmount)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) swap(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.SwapReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.Swap(req.ChainId, req.Amount, req.FromTokenAddress, req.ToTokenAddress, req.UserWalletAddress, req.Slippage, req.SwapReceiverAddress)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) bridgeTokensPairs(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetBridgeTokensPairsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetBridgeTokensPairs(req.FromChainId)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Server) crossChainQuote(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req model.GetCrossChainQuoteReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}
	res, err := s.CoreAPI.OkxDexServiceApi.GetCrossChainQuote(req.FromChainId, req.ToChainId, req.FromTokenAddress, req.ToTokenAddress, req.Amount, req.Slippage)
	if err != nil {
		return nil, err
	}

	return res, nil
}
