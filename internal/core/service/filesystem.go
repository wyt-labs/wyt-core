package service

import (
	"io"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

type FileSystemService struct {
	baseComponent *base.Component
	fileSystemDao *dao.FileSystemDao
}

func NewFileSystemService(baseComponent *base.Component, fileSystemDao *dao.FileSystemDao) (*FileSystemService, error) {
	return &FileSystemService{
		baseComponent: baseComponent,
		fileSystemDao: fileSystemDao,
	}, nil
}

func (s *FileSystemService) Upload(ctx *reqctx.ReqCtx, bucketName string, fileName string, data io.Reader, metadata string) (id string, err error) {
	return s.fileSystemDao.Upload(ctx, bucketName, fileName, data, metadata)
}

func (s *FileSystemService) Download(ctx *reqctx.ReqCtx, bucketName string, id string) (*entity.File, error) {
	return s.fileSystemDao.Download(ctx, bucketName, id)
}
