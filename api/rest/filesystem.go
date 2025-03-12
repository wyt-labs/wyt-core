package rest

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func (s *Server) fsUpload(ctx *reqctx.ReqCtx, c *gin.Context) (any, error) {
	var req entity.FileUploadReq
	if err := c.ShouldBindQuery(&req); err != nil {
		return nil, err
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		return nil, err
	}

	bucketName := req.Type.String()
	ctx.AddCustomLogFields(map[string]any{
		"bucket":    bucketName,
		"file_name": fileHeader.Filename,
		"size":      fileHeader.Size,
	})
	res, err := s.FileSystemService.Upload(ctx, bucketName, fileHeader.Filename, file, "")
	if err != nil {
		return nil, err
	}

	return &entity.FileUploadRes{
		URL: s.generateFileURL(bucketName, res),
	}, nil
}

type FileContentReq struct {
	Bucket string `uri:"bucket"`
	ID     string `uri:"id"`
}

func (s *Server) fsContent(c *gin.Context) {
	err := func() error {
		req := &FileContentReq{}
		if err := c.ShouldBindUri(req); err != nil {
			return err
		}

		res, err := s.FileSystemService.Download(s.baseComponent.BackgroundContext(), req.Bucket, req.ID)
		if err != nil {
			return err
		}

		contentType := mime.TypeByExtension(filepath.Ext(res.Name))
		var extraHeaders map[string]string
		if contentType == "" {
			contentType = "application/octet-stream"
			extraHeaders = map[string]string{
				"Content-Disposition":       "attachment; filename=" + res.Name,
				"Content-Transfer-Encoding": "binary",
				"Cache-Control":             "no-cache",
			}
		}
		c.DataFromReader(http.StatusOK, res.Length, contentType, res.Data, extraHeaders)
		return nil
	}()
	if err != nil {
		code := errcode.DecodeError(err)
		msg := err.Error()
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    code,
			"message": msg,
		})
		return
	}
}

func (s *Server) generateFileURL(bucketName string, id string) string {
	return fmt.Sprintf("http://%s/api/v1/fs/files/%s/%s", s.baseComponent.Config.App.AccessDomain, bucketName, id)
}
