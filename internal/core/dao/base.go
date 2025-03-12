package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

func init() {
	basic.RegisterComponents(NewDB, NewUserDao, NewProjectDao, NewFileSystemDao, NewMiscDao, NewSystemCacheDao, NewChatDao, NewWebsiteDao, NewUserPluginDao)
}

var authMechanisms = []string{
	"SCRAM-SHA-256",
	"SCRAM-SHA-1",
	"PLAIN",
	"MONGODB-CR",
}

type DB struct {
	baseComponent *base.Component
	Client        *mongo.Client
	DB            *mongo.Database
	DBConfig      *config.DB
}

func NewDB(baseComponent *base.Component) *DB {
	d := &DB{
		baseComponent: baseComponent,
		DBConfig:      &baseComponent.Config.DB,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func NewSpecificDB(DBConfig *config.DB) *DB {
	return &DB{
		DBConfig: DBConfig,
	}
}

func (d *DB) retryAuthMechanismConnect() (*mongo.Client, error) {
	opt := options.Client()
	opt.SetMaxPoolSize(uint64(d.DBConfig.Service.MaxPoolSize))
	opt.SetMaxConnecting(uint64(d.DBConfig.Service.MaxPoolSize))
	opt.SetMaxConnIdleTime(d.DBConfig.Service.MaxConnIdleTime.ToDuration())
	opt.ApplyURI(mongoURL(d.DBConfig.Service.IsSrv, d.DBConfig.Service.DBInfo))

	connect := func(mechanism string) (bool, *mongo.Client, error) {
		defaultTimeoutCtx, cancel := context.WithTimeout(context.Background(), d.DBConfig.Service.ConnectTimeout.ToDuration())
		defer cancel()
		if !d.DBConfig.Service.IsSrv {
			if d.DBConfig.Service.Username != "" {
				opt.SetAuth(options.Credential{
					AuthMechanism: mechanism,
					Username:      d.DBConfig.Service.Username,
					Password:      d.DBConfig.Service.Password,
				})
			}
		}

		client, err := mongo.Connect(defaultTimeoutCtx, opt)
		if err != nil {
			return false, nil, err
		}

		defaultTimeoutCtx2, cancel2 := context.WithTimeout(context.Background(), d.DBConfig.Service.ConnectTimeout.ToDuration())
		defer cancel2()
		err = client.Ping(defaultTimeoutCtx2, nil)
		if err == nil {
			return false, client, nil
		}
		_ = client.Disconnect(context.Background())

		if !strings.Contains(err.Error(), "unable to authenticate using mechanism") {
			return false, nil, err
		}
		return true, nil, err
	}
	for _, mechanism := range authMechanisms {
		needRetry, client, err := connect(mechanism)
		if !needRetry {
			return client, err
		}
	}

	return nil, errors.New("connect mongodb failed: unable to authenticate using mechanism [SCRAM-SHA-256, SCRAM-SHA-1, PLAIN, MONGODB-CR], maybe username/pwd/db error")
}

func (d *DB) Start() error {
	client, err := d.retryAuthMechanismConnect()
	if err != nil {
		return err
	}
	d.Client = client
	d.DB = client.Database(d.DBConfig.Service.DBName)
	return nil
}

func (d *DB) Stop() error {
	return d.Client.Disconnect(context.Background())
}

func mongoURL(isSrv bool, dbInfo config.DBInfo) string {
	if isSrv {
		uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority&appName=PumpFun", dbInfo.Username, dbInfo.Password, dbInfo.IP, dbInfo.DBName)
		return uri
	} else {
		if dbInfo.Username == "" {
			return fmt.Sprintf("mongodb://%s:%d/%s", dbInfo.IP, dbInfo.Port, dbInfo.DBName)
		}
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", dbInfo.Username, dbInfo.Password, dbInfo.IP, dbInfo.Port, dbInfo.DBName)
	}
}

func (d *DB) createIndexes(collection *mongo.Collection, isUnique bool, fields []string) error {
	var indexes []mongo.IndexModel
	for _, field := range fields {
		name := "_" + field
		opts := &options.IndexOptions{
			Name:   &name,
			Unique: &isUnique,
		}
		if isUnique {
			opts.SetSparse(true)
		}
		indexes = append(indexes, mongo.IndexModel{
			Keys: bson.D{
				{Key: field, Value: 1},
			},
			Options: opts,
		})
	}
	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return errors.Wrapf(err, "failed to create mongodb index, collection: %v, fields: %v", collection.Name(), fields)
	}
	return nil
}

func (d *DB) insert(collection *mongo.Collection, ctx *reqctx.ReqCtx, v any) (primitive.ObjectID, error) {
	res, err := collection.InsertOne(ctx.Ctx, v)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	id, _ := res.InsertedID.(primitive.ObjectID)
	return id, nil
}

func (d *DB) update(collection *mongo.Collection, ctx *reqctx.ReqCtx, id primitive.ObjectID, v any) error {
	res, err := collection.UpdateByID(ctx.Ctx, id, bson.D{bson.E{Key: "$set", Value: v}})
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 && res.MatchedCount != 1 {
		return fmt.Errorf("internal error: update %s info[%s] failed, whether to find: %v", collection.Name(), id.Hex(), res.MatchedCount == 1)
	}

	return nil
}

func (d *DB) upsert(collection *mongo.Collection, ctx *reqctx.ReqCtx, id any, v any) error {
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateByID(ctx.Ctx, id, bson.D{bson.E{Key: "$set", Value: v}}, opts)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) delete(collection *mongo.Collection, ctx *reqctx.ReqCtx, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := collection.UpdateByID(ctx.Ctx, objID, bson.D{bson.E{Key: "$set", Value: bson.M{
		"is_deleted": true,
	}}})
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("internal error: delete %s info[%s] failed, whether to find: %v", collection.Name(), id, res.MatchedCount == 1)
	}
	return nil
}

func (d *DB) deleteByFilter(collection *mongo.Collection, ctx *reqctx.ReqCtx, filter any) error {
	_, err := collection.UpdateMany(ctx.Ctx, filter, bson.D{bson.E{Key: "$set", Value: bson.M{
		"is_deleted": true,
	}}})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) queryByID(collection *mongo.Collection, ctx *reqctx.ReqCtx, id string, result any) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return collection.FindOne(ctx.Ctx, bson.M{"_id": objID, "is_deleted": false}).Decode(result)
}

func (d *DB) queryByFilter(collection *mongo.Collection, ctx *reqctx.ReqCtx, filter any, result any) error {
	return collection.FindOne(ctx.Ctx, filter).Decode(result)
}

func (d *DB) batchQueryByIDs(collection *mongo.Collection, ctx *reqctx.ReqCtx, ids []string, result any) error {
	if len(ids) == 0 {
		return nil
	}

	var a bson.A
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		a = append(a, objID)
	}
	filter := bson.M{"_id": bson.M{"$in": a}}
	cur, err := collection.Find(ctx.Ctx, filter)
	if err != nil {
		return err
	}
	if err := cur.All(ctx.Ctx, result); err != nil {
		return err
	}
	return nil
}

func (d *DB) pageList(collection *mongo.Collection, ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool, results any) (total int64, err error) {
	if filter == nil {
		filter = bson.M{"is_deleted": false}
	}

	cnt, err := collection.CountDocuments(ctx.Ctx, filter)
	if err != nil {
		return 0, err
	}

	findOptions := options.Find()
	if page != 0 && size != 0 {
		findOptions.SetSkip(int64((page - 1) * size)).SetLimit(int64(size))
	}
	if len(sort) != 0 {
		s := bson.M{}
		for field, isAsc := range sort {
			v := 1
			if !isAsc {
				v = -1
			}
			s[field] = v
		}
		findOptions.SetSort(s)
	} else {
		findOptions.SetSort(bson.M{"update_time": -1})
	}

	cursor, err := collection.Find(ctx.Ctx, filter, findOptions)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = cursor.Close(ctx.Ctx)
	}()
	if err := cursor.All(ctx.Ctx, results); err != nil {
		return 0, err
	}
	return cnt, nil
}

func CollateBatchQueryResult[T model.Model](queryIDs []string, result []T) []T {
	collatedRes := make([]T, len(queryIDs))
	filter := make(map[string]int, len(queryIDs))
	for i, v := range queryIDs {
		filter[v] = i
	}
	for _, e := range result {
		id := e.GetID()
		index, ok := filter[id.Hex()]
		if ok {
			collatedRes[index] = e
		}
	}

	return collatedRes
}

func IDsToObjectIDs(ids []string) ([]primitive.ObjectID, error) {
	var a []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		a = append(a, objID)
	}
	return a, nil
}
