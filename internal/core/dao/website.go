package dao

import (
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	subscribeCollectionName = "subscribe"
)

var (
	ErrSubscribeNotExist = errors.New("subscribe email not exist")
)

type WebsiteDao struct {
	baseComponent       *base.Component
	db                  *DB
	subscribeCollection *mongo.Collection
}

func NewWebsiteDao(baseComponent *base.Component, db *DB) *WebsiteDao {
	d := &WebsiteDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *WebsiteDao) Start() error {
	d.subscribeCollection = d.db.DB.Collection(subscribeCollectionName)
	if err := d.db.createIndexes(d.subscribeCollection, true, []string{"email"}); err != nil {
		return err
	}
	return nil
}

func (d *WebsiteDao) Stop() error {
	return nil
}

func (d *WebsiteDao) SubscribeQueryByEmail(ctx *reqctx.ReqCtx, email string) (*model.Subscribe, error) {
	var res model.Subscribe
	if err := d.subscribeCollection.FindOne(ctx.Ctx, bson.M{"email": email}).Decode(&res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSubscribeNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *WebsiteDao) SubscribeAdd(ctx *reqctx.ReqCtx, e *model.Subscribe) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.subscribeCollection, ctx, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *WebsiteDao) SubscribeUpdate(ctx *reqctx.ReqCtx, e *model.Subscribe) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.upsert(d.subscribeCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}
