package dao

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	systemCacheCollectionName = "system_cache"
)

type SystemCacheInfo struct {
	ID   string `bson:"_id"`
	Data string `bson:"data"`
}

type SystemCacheDao struct {
	baseComponent *base.Component
	db            *DB
	collection    *mongo.Collection
}

func NewSystemCacheDao(baseComponent *base.Component, db *DB) *SystemCacheDao {
	d := &SystemCacheDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *SystemCacheDao) Start() error {
	d.collection = d.db.DB.Collection(systemCacheCollectionName)
	return nil
}

func (d *SystemCacheDao) Stop() error {
	return nil
}

func (d *SystemCacheDao) Put(ctx *reqctx.ReqCtx, id string, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return d.db.upsert(d.collection, ctx, id, &SystemCacheInfo{
		ID:   id,
		Data: string(raw),
	})
}

func (d *SystemCacheDao) Get(ctx *reqctx.ReqCtx, id string, res any) error {
	var e SystemCacheInfo
	err := d.collection.FindOne(ctx.Ctx, bson.M{"_id": id}).Decode(&e)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(e.Data), res)
}

func (d *SystemCacheDao) Has(ctx *reqctx.ReqCtx, id string) (bool, error) {
	cnt, err := d.collection.CountDocuments(ctx.Ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return cnt != 0, nil
}

func (d *SystemCacheDao) Delete(ctx *reqctx.ReqCtx, id string) error {
	_, err := d.collection.DeleteOne(ctx.Ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	return nil
}
