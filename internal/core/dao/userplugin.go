package dao

import (
	"time"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	userPluginCollectionName = "userplugin"
)

type UserPluginDao struct {
	baseComponent *base.Component
	db            *DB
	collection    *mongo.Collection
}

func NewUserPluginDao(baseComponent *base.Component, db *DB) *UserPluginDao {
	up := &UserPluginDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(up)
	return up
}

func (up *UserPluginDao) Start() error {
	// 检查集合是否存在
	collections, err := up.db.DB.ListCollectionNames(up.baseComponent.Ctx, bson.M{"name": userPluginCollectionName})
	if err != nil {
		return err
	}
	// 如果集合不存在，则创建集合
	if len(collections) == 0 {
		err = up.db.DB.CreateCollection(up.baseComponent.Ctx, userPluginCollectionName)
		if err != nil {
			return err
		}
	}
	up.collection = up.db.DB.Collection(userPluginCollectionName)
	return nil
}

func (up *UserPluginDao) Stop() error {
	return nil
}

func (up *UserPluginDao) QueryByUserIdAndPj(ctx *reqctx.ReqCtx, userId, pjId string) (*model.UserPlugin, error) {
	var upgin model.UserPlugin
	err := up.collection.FindOne(ctx.Ctx, bson.M{"user_id": userId, "project_id": pjId}).Decode(&upgin)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			up.baseComponent.Logger.Warn("user plugin pin record not found", "userId", userId, "projectId", pjId)
			return nil, nil
		}
		return nil, err
	}
	return &upgin, nil
}

func (up *UserPluginDao) UserPluginPinList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.UserPlugin, int64, error) {
	var res []*model.UserPlugin
	total, err := up.db.pageList(up.collection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (up *UserPluginDao) UpdateUserPin(ctx *reqctx.ReqCtx, e *model.UserPlugin) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := up.db.update(up.collection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (up *UserPluginDao) InsertUserPin(ctx *reqctx.ReqCtx, e *model.UserPlugin) error {
	objId, err := up.db.insert(up.collection, ctx, e)
	if err != nil {
		return err
	}
	up.baseComponent.Logger.Info("insert user agent pin success", "objId", objId)
	return nil
}
