package coincarp

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/basic"
)

const (
	collectionName  = "coincarp"
	systemCacheID   = "coincarp"
	module          = "coincarp"
	writeCollection = "project_edit"
	errorCacheID    = "coincarp_error"
)

func init() {
	basic.RegisterComponents(New)
}

type Coincarp struct {
	component *base.Component
	readDB    *dao.DB
	writeDB   *dao.DB
	cache     *dao.SystemCacheDao

	miscDao    *dao.MiscDao
	projectDao *dao.ProjectDao
}

func New(baseComponent *base.Component, writeDB *dao.DB, systemCache *dao.SystemCacheDao, miscDao *dao.MiscDao, projectDao *dao.ProjectDao) *Coincarp {
	return &Coincarp{
		component:  baseComponent,
		writeDB:    writeDB,
		readDB:     dao.NewSpecificDB(&baseComponent.Config.ETL.Coincarp),
		cache:      systemCache,
		miscDao:    miscDao,
		projectDao: projectDao,
	}
}

func (c *Coincarp) Start() error {
	_, err := c.component.Cron.AddFunc(c.component.Config.ETL.DataRefreshCron, c.syncReadDB)
	if err != nil {
		c.component.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to add syncReadDB task")
		c.component.ComponentShutdown()
		return nil
	}
	return c.readDB.Start()
}

func (c *Coincarp) Stop() error {
	return c.readDB.Stop()
}

func (c *Coincarp) syncReadDB() {
	// get cache and retry to parse, if err happened, ignore it
	c.readCacheRetry()

	// normal procedure: 1. count # need to update or insert
	// if err happened, return
	collection := c.readDB.DB.Collection(collectionName)
	var count, lastCnt int64
	if err := c.retry(func() (bool, error) {
		var err error
		count, err = collection.CountDocuments(c.component.Ctx, bson.D{})
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return false, nil
			}
			c.component.Logger.WithFields(logrus.Fields{"err": err}).Warn("failed to count documents")
			return true, err
		}
		err = c.cache.Get(c.component.BackgroundContext(), systemCacheID, &lastCnt)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				lastCnt = 0
				return false, nil
			}
			c.component.Logger.WithFields(logrus.Fields{"err": err}).Warn("failed to get system cache")
			return true, err
		}
		return false, nil
	}); err != nil {
		return
	}

	// get the # need update, query the database and update
	needUpdate := count - lastCnt
	c.component.Logger.WithFields(logrus.Fields{"module": module}).
		Infof("load cache successful, get cache count %d, current count %d, need update %d", lastCnt, count, needUpdate)
	if needUpdate == 0 {
		c.component.Logger.WithFields(logrus.Fields{"module": module}).Info("no need to update")
		return
	}
	// if traverse return an error, means no update at all, will return directly
	c.traverse(collection, needUpdate, count)
}

func (c *Coincarp) readCacheRetry() {
	var cacheMsgs []ReadDB
	if err := c.retry(func() (bool, error) {
		if err := c.cache.Get(c.component.BackgroundContext(), errorCacheID, &cacheMsgs); err != nil {
			if err == mongo.ErrNoDocuments {
				return false, nil
			}
			return true, err
		}
		return false, nil
	}); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).Error("failed to get documents from cache")
		return
	}
	if len(cacheMsgs) > 0 {
		c.component.Logger.WithFields(logrus.Fields{"module": module}).Info("find some data in cache, will retrieve them first")
		c.getResAndPut(cacheMsgs)
		c.component.Logger.WithFields(logrus.Fields{"module": module}).Info("retrieve data finished")
	}
}

func (c *Coincarp) traverse(collection *mongo.Collection, needUpdate, totalCnt int64) {
	// traverse read database, if err, return
	var cur *mongo.Cursor
	if err := c.retry(func() (bool, error) {
		var err error
		opt := options.Find()
		if needUpdate != totalCnt {
			opt.SetSkip(needUpdate)
		}
		cur, err = collection.Find(c.component.Ctx, bson.D{}, opt)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return false, nil
			}
			return true, err
		}
		return false, nil
	}); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).
			Error("query read collection failed")
		return
	}

	var res []ReadDB
	if err := cur.All(c.component.Ctx, &res); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).
			Error("decode res to ReadDB failed")
		return
	}
	c.getResAndPut(res)
	if err := c.retry(func() (bool, error) {
		if err := c.cache.Put(c.component.BackgroundContext(), systemCacheID, totalCnt); err != nil {
			return true, err
		}
		return false, nil
	}); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).Error("failed to put totalCnt to cache")
	}
}

func (c *Coincarp) getResAndPut(msgs []ReadDB) {
	var failedList []ReadDB
	for _, singleRes := range msgs {
		if err := c.parseSingleRes(singleRes); err != nil {
			failedList = append(failedList, singleRes)
		}
	}
	// update the failed project list
	if err := c.retry(func() (bool, error) {
		if err := c.cache.Put(c.component.BackgroundContext(), errorCacheID, failedList); err != nil {
			return true, err
		}
		return false, nil
	}); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).Error("fail to put failed project in cache")
	}
}

func (c *Coincarp) parseSingleRes(singleRes ReadDB) error {
	res, exists, err := c.checkProjectExists(singleRes.ProjectName)
	if err != nil {
		return err
	}
	if exists {
		return c.updateProject(res, singleRes)
	}
	return c.addNewProject(singleRes)
}
