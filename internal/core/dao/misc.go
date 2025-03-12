package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/cache"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
)

const (
	entityAliasCollectionName    = "entity_alias"
	chainCollectionName          = "chain"
	trackCollectionName          = "track"
	tagCollectionName            = "tag"
	teamImpressionCollectionName = "team_impression"
	investorCollectionName       = "investor"
)

type MiscDao struct {
	baseComponent            *base.Component
	db                       *DB
	entityAliasCollection    *mongo.Collection
	chainCollection          *mongo.Collection
	trackCollection          *mongo.Collection
	tagCollection            *mongo.Collection
	teamImpressionCollection *mongo.Collection
	investorCollection       *mongo.Collection
}

func NewMiscDao(baseComponent *base.Component, db *DB) *MiscDao {
	d := &MiscDao{
		baseComponent: baseComponent,
		db:            db,
	}
	baseComponent.RegisterLifecycleHook(d)
	return d
}

func (d *MiscDao) Start() error {
	d.entityAliasCollection = d.db.DB.Collection(entityAliasCollectionName)
	d.chainCollection = d.db.DB.Collection(chainCollectionName)
	d.trackCollection = d.db.DB.Collection(trackCollectionName)
	d.tagCollection = d.db.DB.Collection(tagCollectionName)
	d.teamImpressionCollection = d.db.DB.Collection(teamImpressionCollectionName)
	d.investorCollection = d.db.DB.Collection(investorCollectionName)

	if err := d.db.createIndexes(d.chainCollection, true, []string{"name", "name_zh"}); err != nil {
		return err
	}
	if err := d.db.createIndexes(d.trackCollection, true, []string{"name", "name_zh"}); err != nil {
		return err
	}
	if err := d.db.createIndexes(d.tagCollection, true, []string{"name", "name_zh"}); err != nil {
		return err
	}
	if err := d.db.createIndexes(d.teamImpressionCollection, true, []string{"name", "name_zh"}); err != nil {
		return err
	}
	return nil
}

func (d *MiscDao) Stop() error {
	return nil
}

// ----- entity alias -----

func (d *MiscDao) EntityAliasBatchAdd(ctx *reqctx.ReqCtx, elements []*model.EntityAlias) error {
	now := model.JSONTime(time.Now())
	docs := lo.Map(elements, func(item *model.EntityAlias, index int) any {
		item.CreateTime = now
		return item
	})
	_, err := d.entityAliasCollection.InsertMany(ctx.Ctx, docs)
	if err != nil {
		return err
	}
	return nil
}

func (d *MiscDao) EntityAliasBatchQuery(ctx *reqctx.ReqCtx, aliasList []string) ([]*model.EntityAlias, error) {
	var res []*model.EntityAlias

	if len(aliasList) == 0 {
		return res, nil
	}

	var a bson.A
	for _, alias := range aliasList {
		a = append(a, alias)
	}
	cur, err := d.entityAliasCollection.Find(ctx.Ctx, bson.M{"_id": bson.M{"$in": a}})
	if err != nil {
		return nil, err
	}
	if err := cur.All(ctx.Ctx, &res); err != nil {
		return nil, err
	}

	collatedRes := make([]*model.EntityAlias, len(aliasList))
	filter := make(map[string]int, len(aliasList))
	for i, v := range aliasList {
		filter[v] = i
	}
	for _, e := range res {
		index, ok := filter[e.Alias]
		if ok {
			collatedRes[index] = e
		}
	}

	for i, alias := range aliasList {
		if collatedRes[i] == nil {
			return nil, errcode.ErrInvestorNotExist.Wrap(fmt.Sprintf("check entity alias[%s] failed", alias))
		}
	}
	return collatedRes, nil
}

func (d *MiscDao) EntityAliasList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.EntityAlias, int64, error) {
	var res []*model.EntityAlias
	if filter == nil {
		filter = bson.M{}
	}
	if len(sort) == 0 {
		sort = map[string]bool{"create_time": false}
	}

	total, err := d.db.pageList(d.entityAliasCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

// ----- chain -----

func (d *MiscDao) ChainAdd(ctx *reqctx.ReqCtx, e *model.Chain) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.chainCollection, ctx, e)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return errcode.ErrChainAlreadyExist
		}
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, chainCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) ChainUpdate(ctx *reqctx.ReqCtx, e *model.Chain) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.chainCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, chainCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) ChainQuery(ctx *reqctx.ReqCtx, id string) (*model.Chain, error) {
	e, exist := cache.GetFromMemCache[*model.Chain](d.baseComponent.MemCache, chainCollectionName, id)
	if exist {
		return e, nil
	}
	var res model.Chain
	if err := d.db.queryByID(d.chainCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrChainNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *MiscDao) ChainBatchQuery(ctx *reqctx.ReqCtx, ids []string) ([]*model.Chain, error) {
	var res []*model.Chain
	for _, id := range ids {
		e, err := d.ChainQuery(ctx, id)
		if err != nil && err != errcode.ErrChainNotExist {
			return nil, err
		}
		if err == errcode.ErrChainNotExist {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (d *MiscDao) ChainBatchCheck(ctx *reqctx.ReqCtx, ids []string) error {
	for _, id := range ids {
		_, err := d.ChainQuery(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "check chain[%s] failed", id)
		}
	}
	return nil
}

func (d *MiscDao) ChainList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.Chain, int64, error) {
	var res []*model.Chain
	total, err := d.db.pageList(d.chainCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

// ----- track -----

func (d *MiscDao) TrackAdd(ctx *reqctx.ReqCtx, e *model.Track) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.trackCollection, ctx, e)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return errcode.ErrTrackAlreadyExist
		}
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, trackCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TrackUpdate(ctx *reqctx.ReqCtx, e *model.Track) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.trackCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, trackCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TrackQuery(ctx *reqctx.ReqCtx, id string) (*model.Track, error) {
	e, exist := cache.GetFromMemCache[*model.Track](d.baseComponent.MemCache, trackCollectionName, id)
	if exist {
		return e, nil
	}
	var res model.Track
	if err := d.db.queryByID(d.trackCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrTrackNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *MiscDao) TrackBatchQuery(ctx *reqctx.ReqCtx, ids []string) ([]*model.Track, error) {
	var res []*model.Track
	for _, id := range ids {
		e, err := d.TrackQuery(ctx, id)
		if err != nil && err != errcode.ErrTrackNotExist {
			return nil, err
		}
		if err == errcode.ErrTrackNotExist {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (d *MiscDao) TrackBatchCheck(ctx *reqctx.ReqCtx, ids []string) error {
	for _, id := range ids {
		_, err := d.TrackQuery(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "check track[%s] failed", id)
		}
	}
	return nil
}

func (d *MiscDao) TrackList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.Track, int64, error) {
	var res []*model.Track
	total, err := d.db.pageList(d.trackCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

// ----- tag -----

func (d *MiscDao) TagAdd(ctx *reqctx.ReqCtx, e *model.Tag) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.tagCollection, ctx, e)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return errcode.ErrTagAlreadyExist
		}
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, tagCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TagUpdate(ctx *reqctx.ReqCtx, e *model.Tag) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.tagCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, tagCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TagQuery(ctx *reqctx.ReqCtx, id string) (*model.Tag, error) {
	e, exist := cache.GetFromMemCache[*model.Tag](d.baseComponent.MemCache, tagCollectionName, id)
	if exist {
		return e, nil
	}
	var res model.Tag
	if err := d.db.queryByID(d.tagCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrTagNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *MiscDao) TagBatchQuery(ctx *reqctx.ReqCtx, ids []string) ([]*model.Tag, error) {
	var res []*model.Tag
	for _, id := range ids {
		e, err := d.TagQuery(ctx, id)
		if err != nil && err != errcode.ErrTagNotExist {
			return nil, err
		}
		if err == errcode.ErrTagNotExist {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (d *MiscDao) TagBatchCheck(ctx *reqctx.ReqCtx, ids []string) error {
	for _, id := range ids {
		_, err := d.TagQuery(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "check tag[%s] failed", id)
		}
	}
	return nil
}

func (d *MiscDao) TagList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.Tag, int64, error) {
	var res []*model.Tag
	total, err := d.db.pageList(d.tagCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

// ----- team impressions -----

func (d *MiscDao) TeamImpressionAdd(ctx *reqctx.ReqCtx, e *model.TeamImpression) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.teamImpressionCollection, ctx, e)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return errcode.ErrTeamImpressionAlreadyExist
		}
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, teamImpressionCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TeamImpressionUpdate(ctx *reqctx.ReqCtx, e *model.TeamImpression) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.teamImpressionCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	cache.PutToMemCache(d.baseComponent.MemCache, teamImpressionCollectionName, e.ID.Hex(), e)
	return nil
}

func (d *MiscDao) TeamImpressionQuery(ctx *reqctx.ReqCtx, id string) (*model.TeamImpression, error) {
	e, exist := cache.GetFromMemCache[*model.TeamImpression](d.baseComponent.MemCache, teamImpressionCollectionName, id)
	if exist {
		return e, nil
	}
	var res model.TeamImpression
	if err := d.db.queryByID(d.teamImpressionCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrTeamImpressionNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *MiscDao) TeamImpressionBatchQuery(ctx *reqctx.ReqCtx, ids []string) ([]*model.TeamImpression, error) {
	var res []*model.TeamImpression
	for _, id := range ids {
		e, err := d.TeamImpressionQuery(ctx, id)
		if err != nil && err != errcode.ErrTeamImpressionNotExist {
			return nil, err
		}
		res = append(res, e)
	}
	return res, nil
}

func (d *MiscDao) TeamImpressionBatchCheck(ctx *reqctx.ReqCtx, ids []string) error {
	for _, id := range ids {
		_, err := d.TeamImpressionQuery(ctx, id)
		if err != nil {
			return errors.Wrapf(err, "check team-impression[%s] failed", id)
		}
	}
	return nil
}

func (d *MiscDao) TeamImpressionList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.TeamImpression, int64, error) {
	var res []*model.TeamImpression
	total, err := d.db.pageList(d.teamImpressionCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}

// ----- investor -----

func (d *MiscDao) InvestorAdd(ctx *reqctx.ReqCtx, e *model.Investor) error {
	var err error
	e.BaseModel, err = model.NewBaseModel(ctx.Caller)
	if err != nil {
		return err
	}
	e.ID, err = d.db.insert(d.investorCollection, ctx, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *MiscDao) InvestorUpdate(ctx *reqctx.ReqCtx, e *model.Investor) error {
	e.UpdateTime = model.JSONTime(time.Now())
	err := d.db.update(d.investorCollection, ctx, e.ID, e)
	if err != nil {
		return err
	}
	return nil
}

func (d *MiscDao) InvestorQuery(ctx *reqctx.ReqCtx, id string) (*model.Investor, error) {
	var res model.Investor
	if err := d.db.queryByID(d.investorCollection, ctx, id, &res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errcode.ErrInvestorNotExist
		}
		return nil, err
	}
	return &res, nil
}

func (d *MiscDao) InvestorBatchQuery(ctx *reqctx.ReqCtx, ids []string) ([]*model.Investor, error) {
	var res []*model.Investor
	err := d.db.batchQueryByIDs(d.investorCollection, ctx, ids, &res)
	if err != nil {
		return nil, err
	}
	res = CollateBatchQueryResult(ids, res)
	for i, id := range ids {
		if res[i] == nil {
			return nil, errcode.ErrInvestorNotExist.Wrap(fmt.Sprintf("find investor[%s] failed", id))
		}
	}
	return res, nil
}

func (d *MiscDao) InvestorBatchCheck(ctx *reqctx.ReqCtx, ids []string) error {
	var res []*model.Investor
	err := d.db.batchQueryByIDs(d.investorCollection, ctx, ids, &res)
	if err != nil {
		return err
	}

	res = CollateBatchQueryResult(ids, res)
	for i, id := range ids {
		if res[i] == nil {
			return errcode.ErrInvestorNotExist.Wrap(fmt.Sprintf("check investor[%s] failed", id))
		}
	}
	return nil
}

func (d *MiscDao) InvestorList(ctx *reqctx.ReqCtx, page uint64, size uint64, filter any, sort map[string]bool) ([]*model.Investor, int64, error) {
	var res []*model.Investor
	total, err := d.db.pageList(d.investorCollection, ctx, page, size, filter, sort, &res)
	if err != nil {
		return nil, 0, err
	}
	return res, total, nil
}
