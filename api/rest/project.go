package rest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) projectInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectInfoReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.Info(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) projectSimpleInfo(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectSimpleInfoReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("id", req.ID)

	res, err := s.ProjectService.SimpleInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) projectList(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectListReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("page", req.Page)
	ctx.AddCustomLogField("size", req.Size)
	ctx.AddCustomLogFieldOnError("query", req.Query)
	conditions, _ := json.Marshal(req.Conditions)
	ctx.AddCustomLogFieldOnError("conditions", string(conditions))
	ctx.AddCustomLogFieldOnError("sort_field", req.SortField)
	ctx.AddCustomLogFieldOnError("is_asc", req.Page)

	res, err := s.ProjectService.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) projectInfoCompare(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectInfoCompareReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	req.ProjectIDs = strings.TrimSpace(req.ProjectIDs)
	if req.ProjectIDs == "" {
		return nil, errcode.ErrRequestParameter.Wrap("project-ids cannot be empty")
	}
	projectIDs := strings.Split(req.ProjectIDs, ",")
	if len(projectIDs) > 3 {
		return nil, errcode.ErrRequestParameter.Wrap("the maximum number of project-ids is three")
	}
	for i, id := range projectIDs {
		if id == "" {
			return nil, errcode.ErrRequestParameter.Wrap(fmt.Sprintf("the project-id[idx:%v] cannot be empty", i))
		}
	}

	req.DecodedProjectIDs = projectIDs

	return s.ProjectService.InfoCompare(ctx, req)
}

func (s *Server) projectMetricsCompare(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	req := &entity.ProjectMetricsCompareReq{}
	if err := c.ShouldBindQuery(req); err != nil {
		return nil, err
	}
	ctx.AddCustomLogField("projects", req.ProjectIDs)
	ctx.AddCustomLogField("type", req.Type)
	ctx.AddCustomLogField("start", req.StartTime)
	ctx.AddCustomLogField("end", req.EndTime)
	ctx.AddCustomLogField("interval", req.Interval)

	req.ProjectIDs = strings.TrimSpace(req.ProjectIDs)
	if req.ProjectIDs == "" {
		return nil, errcode.ErrRequestParameter.Wrap("project-ids cannot be empty")
	}
	projectIDs := strings.Split(req.ProjectIDs, ",")
	if len(projectIDs) > 3 {
		return nil, errcode.ErrRequestParameter.Wrap("the maximum number of project-ids is three")
	}
	for i, id := range projectIDs {
		if id == "" {
			return nil, errcode.ErrRequestParameter.Wrap(fmt.Sprintf("the project-id[idx:%v] cannot be empty", i))
		}
	}
	req.DecodedProjectIDs = projectIDs

	if req.StartTime > req.EndTime {
		return nil, errcode.ErrRequestParameter.Wrap("end_time must be greater than start_time")
	}

	return s.ProjectService.MetricsCompare(ctx, req)
}
