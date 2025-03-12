package dao

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	userCollectionName = "user"
)

type UserDao struct {
	baseComponent *base.Component
	db            *DB
	collection    *mongo.Collection
}

func NewUserDao(baseComponent *base.Component, db *DB) *UserDao {
	d := &UserDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *UserDao) Start() error {
	d.collection = d.db.DB.Collection(userCollectionName)
	if err := d.db.createIndexes(d.collection, true, []string{"addr"}); err != nil {
		return err
	}
	return nil
}

func (d *UserDao) Stop() error {
	return nil
}

func (d *UserDao) Create(ctx *reqctx.ReqCtx, user *model.User) error {
	if user.Addr != "" && user.Addr == d.baseComponent.Config.App.AdminAddr {
		user.Role = model.UserRoleAdmin
	}
	var err error
	user.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	user.ID, err = d.db.insert(d.collection, ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (d *UserDao) QueryByID(ctx *reqctx.ReqCtx, id string) (*model.User, error) {
	var user model.User
	if err := d.db.queryByID(d.collection, ctx, id, &user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrUserNotExist
		}
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) QueryByAddr(ctx *reqctx.ReqCtx, addr string) (*model.User, error) {
	var user model.User
	err := d.collection.FindOne(ctx.Ctx, bson.M{"addr": addr}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrUserNotExist
		}
		return nil, err
	}
	return &user, nil
}

func (d *UserDao) HasByAddr(ctx *reqctx.ReqCtx, addr string) (bool, error) {
	cnt, err := d.collection.CountDocuments(ctx.Ctx, bson.M{"addr": addr})
	if err != nil {
		return false, err
	}
	return cnt != 0, nil
}

func (d *UserDao) Update(ctx *reqctx.ReqCtx, user *model.User) error {
	user.UpdateTime = model.JSONTime(time.Now())

	return d.db.update(d.collection, ctx, user.ID, user)
}

func (d *UserDao) Insert(ctx *reqctx.ReqCtx, user *model.User) error {
	var err error
	user.ID, err = d.db.insert(d.collection, ctx, user)
	if err != nil {
		return err
	}
	return nil
}
