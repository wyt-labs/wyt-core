package dao

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	chatWindowCollectionName  = "chat_window"
	chatHistoryCollectionName = "chat_history"
)

type ChatDao struct {
	baseComponent         *base.Component
	db                    *DB
	chatWindowCollection  *mongo.Collection
	chatHistoryCollection *mongo.Collection
}

func NewChatDao(baseComponent *base.Component, db *DB) *ChatDao {
	d := &ChatDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *ChatDao) Start() error {
	d.chatWindowCollection = d.db.DB.Collection(chatWindowCollectionName)
	d.chatHistoryCollection = d.db.DB.Collection(chatHistoryCollectionName)

	if err := d.db.createIndexes(d.chatHistoryCollection, false, []string{"window_id", "index"}); err != nil {
		return err
	}
	return nil
}

func (d *ChatDao) Stop() error {
	return nil
}

func (d *ChatDao) ChatWindowAdd(ctx *reqctx.ReqCtx, e *model.ChatWindow) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.chatWindowCollection, ctx, e)
	if err != nil {
		return err
	}
	return err
}

func (d *ChatDao) ChatWindowQuery(ctx *reqctx.ReqCtx, id string) (*model.ChatWindow, error) {
	var res model.ChatWindow
	if err := d.db.queryByID(d.chatWindowCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrChatWindowNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *ChatDao) ChatWindowList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.ChatWindow, int64, error) {
	var res []*model.ChatWindow
	total, err := d.db.pageList(d.chatWindowCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (d *ChatDao) ChatWindowUpdate(ctx *reqctx.ReqCtx, e *model.ChatWindow) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.chatWindowCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *ChatDao) ChatWindowDelete(ctx *reqctx.ReqCtx, id string) error {
	if err := d.db.delete(d.chatWindowCollection, ctx, id); err != nil {
		return err
	}
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return d.db.deleteByFilter(d.chatHistoryCollection, ctx, bson.M{"window_id": objID})
}

func (d *ChatDao) ChatHistoryAdd(ctx *reqctx.ReqCtx, e *model.ChatHistory) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.chatHistoryCollection, ctx, e)
	if err != nil {
		return err
	}
	return err
}

func (d *ChatDao) ChatHistoryQuery(ctx *reqctx.ReqCtx, windowID primitive.ObjectID, index uint64) (*model.ChatHistory, error) {
	var res model.ChatHistory
	if err := d.db.queryByFilter(d.chatHistoryCollection, ctx, bson.M{"window_id": windowID, "index": index}, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *ChatDao) ChatHistoryList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.ChatHistory, int64, error) {
	var res []*model.ChatHistory
	total, err := d.db.pageList(d.chatHistoryCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}
