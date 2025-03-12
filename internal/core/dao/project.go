package dao

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/entity"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	projectEditCollectionName = "project_edit"
	projectViewCollectionName = "project_view"
)

type ProjectDao struct {
	baseComponent  *base.Component
	db             *DB
	editCollection *mongo.Collection
	viewCollection *mongo.Collection
}

func NewProjectDao(baseComponent *base.Component, db *DB) *ProjectDao {
	d := &ProjectDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *ProjectDao) Start() error {
	d.editCollection = d.db.DB.Collection(projectEditCollectionName)
	d.viewCollection = d.db.DB.Collection(projectViewCollectionName)

	if err := d.db.createIndexes(d.editCollection, false, []string{"basic.name", "tokenomics.token_symbol"}); err != nil {
		return err
	}
	if err := d.db.createIndexes(d.viewCollection, false, []string{"basic.name", "tokenomics.token_symbol"}); err != nil {
		return err
	}
	return nil
}

func (d *ProjectDao) Stop() error {
	return nil
}

func (d *ProjectDao) selectCollection(isView bool) *mongo.Collection {
	if isView {
		return d.viewCollection
	}
	return d.editCollection
}

func (d *ProjectDao) Add(ctx *reqctx.ReqCtx, isView bool, e *model.Project) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.CalculateDerivedData(d.baseComponent.Config)
	e.ID, err = d.db.insert(d.selectCollection(isView), ctx, e)
	if err != nil {
		return err
	}
	return err
}

func (d *ProjectDao) Update(ctx *reqctx.ReqCtx, isView bool, e *model.Project) error {
	e.UpdateTime = model.JSONTime(time.Now())
	e.CalculateDerivedData(d.baseComponent.Config)
	err := d.db.update(d.selectCollection(isView), ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectDao) UpdateInternalData(ctx *reqctx.ReqCtx, isView bool, e *model.Project) error {
	e.CalculateDerivedData(d.baseComponent.Config)
	err := d.db.update(d.selectCollection(isView), ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectDao) Upsert(ctx *reqctx.ReqCtx, isView bool, e *model.Project) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.upsert(d.selectCollection(isView), ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectDao) Query(ctx *reqctx.ReqCtx, isView bool, id string) (*model.Project, error) {
	var res model.Project
	if err := d.db.queryByID(d.selectCollection(isView), ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrProjectNotExist.Wrap("not found project: " + id)
		}
		return nil, err
	}
	return &res, nil
}

func (d *ProjectDao) QueryByProjectName(ctx *reqctx.ReqCtx, isView bool, projectName string) (*model.Project, error) {
	var res model.Project
	if err := d.db.queryByFilter(d.selectCollection(isView), ctx, bson.D{{Key: "basic.name", Value: projectName}}, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrProjectNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *ProjectDao) ExistsByProjectName(ctx *reqctx.ReqCtx, isView bool, projectName string) (bool, error) {
	cnt, err := d.selectCollection(isView).CountDocuments(ctx.Ctx, bson.M{"basic.name": projectName, "is_deleted": false})
	if err != nil {
		return false, err
	}
	return cnt != 0, nil
}

func (d *ProjectDao) Delete(ctx *reqctx.ReqCtx, isView bool, id string) error {
	return d.db.delete(d.selectCollection(isView), ctx, id)
}

func (d *ProjectDao) BatchQuery(ctx *reqctx.ReqCtx, isView bool, ids []string) ([]*model.Project, error) {
	var res []*model.Project
	for _, id := range ids {
		e, err := d.Query(ctx, isView, id)
		if err != nil {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func (d *ProjectDao) CustomBatchQuery(ctx *reqctx.ReqCtx, isView bool, ids []string, res any) error {
	return d.db.batchQueryByIDs(d.selectCollection(isView), ctx, ids, res)
}

func (d *ProjectDao) BatchCheck(ctx *reqctx.ReqCtx, isView bool, ids []string) error {
	var list []*entity.ProjectSimpleOutputQueryWrapper
	if err := d.CustomBatchQuery(ctx, isView, ids, &list); err != nil {
		return err
	}
	list = CollateBatchQueryResult(ids, list)

	for _, v := range list {
		if v == nil {
			return errcode.ErrProjectNotExist.Wrap(fmt.Sprintf("top project[%s] not found", v.ID.Hex()))
		}
		if v.Status == model.ProjectStatusIndexed {
			return errors.Errorf("related project[%s] not published", v.ID.Hex())
		}
	}
	return nil
}

func (d *ProjectDao) List(ctx *reqctx.ReqCtx, isView bool, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.Project, int64, error) {
	var res []*model.Project
	total, err := d.db.pageList(d.selectCollection(isView), ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

func (d *ProjectDao) CustomList(ctx *reqctx.ReqCtx, isView bool, page uint64, size uint64, filter any, sort map[string]bool, res any) (int64, error) {
	total, err := d.db.pageList(d.selectCollection(isView), ctx, page, size, filter, sort, res)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (d *ProjectDao) CustomQuery(ctx *reqctx.ReqCtx, isView bool, id string, res any) error {
	if err := d.db.queryByID(d.selectCollection(isView), ctx, id, res); err != nil {
		if err == mongo.ErrNoDocuments {
			return errcode.ErrProjectNotExist
		}
		return err
	}
	return nil
}

func (d *ProjectDao) BatchSearch(ctx *reqctx.ReqCtx, isView bool, keys []string) ([]*model.Project, error) {
	var res []*model.Project
	for _, key := range keys {
		var e model.Project

		err := d.selectCollection(isView).FindOne(ctx.Ctx, bson.M{
			"tokenomics.token_symbol": bson.M{"$regex": fmt.Sprintf("^%s$", key), "$options": "i"},
		}).Decode(&e)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.Wrapf(err, "failed to find project by key [%s]", key)
			}
		} else {
			res = append(res, &e)
			continue
		}

		err = d.selectCollection(isView).FindOne(ctx.Ctx, bson.M{
			"basic.name": bson.M{"$regex": fmt.Sprintf("^%s$", key), "$options": "i"},
		}).Decode(&e)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return nil, errors.Wrapf(err, "failed to find project by key [%s]", key)
			}
		} else {
			res = append(res, &e)
			continue
		}

		if err := d.selectCollection(isView).FindOne(ctx.Ctx, bson.M{
			"basic.name": bson.M{"$regex": key, "$options": "i"},
		}).Decode(&e); err != nil {
			return nil, errors.Wrapf(err, "failed to find project by key [%s]", key)
		}
		res = append(res, &e)
	}
	return res, nil
}
