package coincarp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/reqctx"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

const (
	readingDBName   = "reading"
	writingDBName   = "writing"
	projectEditName = "project_edit"

	testName1 = "HUG"
	testName2 = "FACE"
)

func generateDate(name string) ReadDB {
	id, _ := uuid.NewUUID()
	return ReadDB{
		ID:          id.String(),
		ProjectCode: "hug",
		ProjectName: name,
		Logo:        "/logo/project/hug.png?style=36&v=1681440588",
		CategoryList: []Category{
			{
				Code: "nfts",
				Name: "NFTs",
			},
		},
		FundCode:      "hug-seed",
		FundStageCode: "seed",
		FundStageName: "种子轮",
		FundAmount:    5000000,
		Valulation:    0,
		FundDate:      1681257600,
		InvestorCodes: "digital,okx,bold",
		InvestorNames: "DIGITAL,OKX,BOLD",
		InvestorLogos: "1,1,1",
		InvestorCount: 8,
		InvestorList: []Investor{
			{
				InvestorLogo: "/logo/investor/digital.png?style=36",
				InvestorCode: "digital",
				InvestorName: "DIGITAL",
			}, {
				InvestorLogo: "/logo/investor/okx.png?style=36",
				InvestorCode: "okx",
				InvestorName: "OKX",
			}, {
				InvestorLogo: "/logo/investor/bold.png?style=36",
				InvestorCode: "bold",
				InvestorName: "BOLD",
			},
		},
	}
}

func newCoincarp(t *testing.T, port int) *Coincarp {
	component := base.NewMockBaseComponent(t)
	component.Config.DB.Service = config.Mongodb{
		DBInfo: config.DBInfo{
			IP:       "127.0.0.1",
			Port:     uint32(port),
			Username: "",
			Password: "",
			DBName:   writingDBName,
		},
		ConnectTimeout:  config.Duration(3 * time.Second),
		MaxPoolSize:     200,
		MaxConnIdleTime: config.Duration(30 * time.Minute),
	}
	component.Config.ETL.Coincarp.Service = config.Mongodb{
		DBInfo: config.DBInfo{
			IP:       "127.0.0.1",
			Port:     uint32(port),
			Username: "",
			Password: "",
			DBName:   readingDBName,
		},
		ConnectTimeout:  config.Duration(3 * time.Second),
		MaxPoolSize:     200,
		MaxConnIdleTime: config.Duration(30 * time.Minute),
	}
	component.Config.ETL.DataRefreshCron = "1 0 * * *"
	db := dao.NewDB(component)
	systemCache := dao.NewSystemCacheDao(component, db)
	miscDao := dao.NewMiscDao(component, db)
	projectDao := dao.NewProjectDao(component, db)
	coinCarp := New(component, db, systemCache, miscDao, projectDao)
	return coinCarp
}

func prepareReadData(readDB *dao.DB, name string) error {
	collection := readDB.DB.Collection(collectionName)
	_, err := collection.InsertOne(context.Background(), generateDate(name))
	return err
}

func prepareCacheData(ctx *reqctx.ReqCtx, cache *dao.SystemCacheDao, name string) error {
	return cache.Put(ctx, errorCacheID, []ReadDB{generateDate(name)})
}

func prepareWriteData(writeDB *dao.DB) error {
	collection := writeDB.DB.Collection(projectEditName)
	project := model.Project{
		Basic: model.ProjectBasic{
			Name: "HUG",
		},
	}
	_, err := collection.InsertOne(context.Background(), project)
	return err
}

func start(coincarp *Coincarp) error {
	if err := coincarp.writeDB.Start(); err != nil {
		return err
	}
	if err := coincarp.miscDao.Start(); err != nil {
		return err
	}
	if err := coincarp.projectDao.Start(); err != nil {
		return err
	}
	if err := coincarp.cache.Start(); err != nil {
		return err
	}
	return coincarp.Start()
}

func stop(coincarp *Coincarp) error {
	if err := coincarp.cache.Stop(); err != nil {
		return err
	}
	if err := coincarp.projectDao.Stop(); err != nil {
		return err
	}
	if err := coincarp.miscDao.Stop(); err != nil {
		return err
	}
	if err := coincarp.writeDB.Stop(); err != nil {
		return err
	}
	return coincarp.Stop()
}

func TestAddNewProject(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	if err := prepareReadData(coincarp.readDB, testName1); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "HUG", res[0].Basic.Name)
	var count int
	if err = coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, count)
}

func TestUpdate(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	if err := prepareReadData(coincarp.readDB, testName1); err != nil {
		t.Fatal(err)
	}
	if err := prepareWriteData(coincarp.writeDB); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0].Funding.FundingDetails))
	assert.Equal(t, "seed", res[0].Funding.FundingDetails[0].Round)
	var count int
	if err = coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, count)
}

func TestRetrieveAddCache(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	if err := prepareCacheData(coincarp.component.BackgroundContext(), coincarp.cache, testName1); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0].Funding.FundingDetails))
	assert.Equal(t, "seed", res[0].Funding.FundingDetails[0].Round)
	var count int
	err = coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count)
	assert.EqualError(t, err, mongo.ErrNoDocuments.Error())
}

func TestRetrieveUpdateCache(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	if err := prepareCacheData(coincarp.component.BackgroundContext(), coincarp.cache, testName1); err != nil {
		t.Fatal(err)
	}
	if err := prepareWriteData(coincarp.writeDB); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0].Funding.FundingDetails))
	assert.Equal(t, "seed", res[0].Funding.FundingDetails[0].Round)
	var count int
	err = coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count)
	assert.EqualError(t, err, mongo.ErrNoDocuments.Error())
}

func TestNoUpdateSkip(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	if err := prepareReadData(coincarp.readDB, testName1); err != nil {
		t.Fatal(err)
	}
	if err = coincarp.cache.Put(coincarp.component.BackgroundContext(), systemCacheID, 1); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(res))
}

func TestAddTwoTimes(t *testing.T) {
	mongoServer, err := util.MockMongoServer()
	defer mongoServer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	coincarp := newCoincarp(t, mongoServer.Port())
	if err := start(coincarp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stop(coincarp); err != nil {
			t.Fatal(err)
		}
	}()
	fmt.Println("1. add 1st time")
	if err := prepareReadData(coincarp.readDB, testName1); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err := coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, len(res))
	assert.Equal(t, 1, len(res[0].Funding.FundingDetails))
	assert.Equal(t, "seed", res[0].Funding.FundingDetails[0].Round)
	var count int
	if err := coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, count)

	fmt.Println("2. add 2nd time")
	if err := prepareReadData(coincarp.readDB, testName2); err != nil {
		t.Fatal(err)
	}
	coincarp.syncReadDB()
	res, _, err = coincarp.projectDao.List(coincarp.component.BackgroundContext(), false, 0, 0, nil, map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, len(res))
	assert.Equal(t, 1, len(res[0].Funding.FundingDetails))
	assert.Equal(t, "seed", res[0].Funding.FundingDetails[0].Round)
	if err = coincarp.cache.Get(coincarp.component.BackgroundContext(), systemCacheID, &count); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, count)
}
