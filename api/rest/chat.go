package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) chatCreate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatCreateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	res, err := s.ChatService.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) agentList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.AgentListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}

	res, err := s.ChatService.AgentList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) pinAgent(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.AgentPinReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.ProjectId == "" {
		return nil, errcode.ErrUserPinPjNotExist
	}
	err := s.ChatService.PinAgent(ctx, req)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) unpinAgent(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.AgentUnPinReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	if req.ProjectId == "" {
		return nil, errcode.ErrUserPinPjNotExist
	}
	err := s.ChatService.UnpinAgent(ctx, req)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) chatList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatListReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)

	res, err := s.ChatService.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatHistory(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatHistoryReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)

	res, err := s.ChatService.History(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatCompletions(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatCompletionsReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ChatService.Completions(ctx, req, false)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatCompletionsJsonMode(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatCompletionsReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ChatService.Completions(ctx, req, true)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatUpdate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)
	ctx.AddCustomLogField("title", req.Title)

	res, err := s.ChatService.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatDelete(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatDeleteReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ChatService.Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) chatDeleteAll(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ChatDeleteAllReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	res, err := s.ChatService.DeleteAll(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
