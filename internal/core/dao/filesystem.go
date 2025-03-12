package dao

import (
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

type Metadata struct {
	Content string `json:"content" bson:"content"`
}

type FileSystemDao struct {
	baseComponent *base.Component
	db            *DB
	fsDB          *mongo.Database
}

func NewFileSystemDao(baseComponent *base.Component, db *DB) *FileSystemDao {
	d := &FileSystemDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *FileSystemDao) Start() error {
	d.fsDB = d.db.Client.Database(fmt.Sprintf("%s-%s", d.baseComponent.Config.DB.Service.DBName, "fs"))

	return nil
}

func (d *FileSystemDao) Stop() error {
	return nil
}

func (d *FileSystemDao) getBucket(bucketName string) (*gridfs.Bucket, error) {
	bucketOptions := options.GridFSBucket().SetName(bucketName)
	return gridfs.NewBucket(d.fsDB, bucketOptions)
}

func (d *FileSystemDao) Upload(ctx *reqctx.ReqCtx, bucketName string, fileName string, data io.Reader, metadata string) (id string, err error) {
	bucket, err := d.getBucket(bucketName)
	if err != nil {
		return "", err
	}
	id = d.baseComponent.UUIDGenerator.Generate().Base58()

	opts := options.GridFSUpload().SetMetadata(&Metadata{
		Content: metadata,
	})
	if err := bucket.UploadFromStreamWithID(id, fileName, data, opts); err != nil {
		return "", err
	}
	return id, nil
}

func (d *FileSystemDao) UploadWithID(ctx *reqctx.ReqCtx, id string, bucketName string, fileName string, data io.Reader, metadata string) (err error) {
	bucket, err := d.getBucket(bucketName)
	if err != nil {
		return err
	}

	opts := options.GridFSUpload().SetMetadata(&Metadata{
		Content: metadata,
	})
	if err := bucket.Delete(id); err != nil && err != gridfs.ErrFileNotFound {
		return err
	}
	if err := bucket.UploadFromStreamWithID(id, fileName, data, opts); err != nil {
		return err
	}
	return nil
}

func (d *FileSystemDao) Download(ctx *reqctx.ReqCtx, bucketName string, id string) (*entity.File, error) {
	bucket, err := d.getBucket(bucketName)
	if err != nil {
		return nil, err
	}

	ds, err := bucket.OpenDownloadStream(id)
	if err != nil {
		return nil, err
	}
	f := ds.GetFile()
	id, _ = f.ID.(string)
	var metadata Metadata
	if err := bson.Unmarshal(f.Metadata, &metadata); err != nil {
		_ = ds.Close()
		return nil, err
	}

	return &entity.File{
		ID:         id,
		Name:       f.Name,
		Length:     f.Length,
		UploadDate: f.UploadDate,
		Metadata:   metadata.Content,
		Data:       ds,
	}, nil
}
