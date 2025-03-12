package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) adminProjectAdd(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectAddReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("name", req.Basic.Name)

	res, err := s.ProjectService.AdminAdd(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectUpdate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)
	ctx.AddCustomLogField("name", req.Basic.Name)

	res, err := s.ProjectService.AdminUpdate(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectSimpleUpdate(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectSimpleUpdateReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)
	ctx.AddCustomLogField("name", req.Name)

	res, err := s.ProjectService.AdminSimpleUpdate(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectPublish(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectPublishReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.AdminPublish(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectAdminInfoReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.AdminInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectSimpleInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectAdminSimpleInfoReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.AdminSimpleInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectAdminListReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)
	ctx.AddCustomLogFieldOnError("query", req.Query)
	ctx.AddCustomLogFieldOnError("status", req.Status)
	ctx.AddCustomLogFieldOnError("sort_field", req.SortField)
	ctx.AddCustomLogFieldOnError("is_asc", req.Page)

	res, err := s.ProjectService.AdminList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectDelete(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectDeleteReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.AdminDelete(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) adminProjectCalculateDerivedData(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectCalculateDerivedDataReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	res, err := s.ProjectService.AdminCalculateDerivedData(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
