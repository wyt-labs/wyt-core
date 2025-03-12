package coincarp

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/core/model"
	"github.com/wyt-labs/wyt-core/internal/pkg/errcode"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

func (c *Coincarp) retry(worker func() (bool, error)) error {
	return util.Retry(c.component.Config.App.RetryInterval.ToDuration(), c.component.Config.App.RetryTime, worker)
}

func (c *Coincarp) checkProjectExists(projectName string) (*model.Project, bool, error) {
	var exists bool
	res := &model.Project{}
	if err := c.retry(func() (bool, error) {
		var err error
		res, err = c.projectDao.QueryByProjectName(c.component.BackgroundContext(), false, projectName)
		switch err {
		case errcode.ErrProjectNotExist:
			exists = false
			return false, nil
		case nil:
			exists = true
			return false, nil
		default:
			c.component.Logger.WithFields(logrus.Fields{"err": err}).Warnf("find record: [%s] failed", projectName)
			return true, err
		}
	}); err != nil {
		c.component.Logger.WithFields(logrus.Fields{"err": err}).Errorf("find record: [%s] failed", projectName)
		return nil, exists, err
	}
	return res, exists, nil
}

func (c *Coincarp) parseFundingDetail(singleRes ReadDB) model.ProjectFundingDetail {
	tm := time.Unix(singleRes.FundDate, 0).Format("2006-01-02")
	return model.ProjectFundingDetail{
		Round:         singleRes.FundStageCode,
		Date:          tm,
		Amount:        uint64(singleRes.FundAmount),
		Valuation:     uint64(singleRes.Valulation),
		Investors:     singleRes.InvestorNames,
		LeadInvestors: "",
		InternalDate:  model.JSONTime{},
	}
}

func (c *Coincarp) updateProject(current *model.Project, singleRes ReadDB) error {
	fundingDetail := c.parseFundingDetail(singleRes)
	current.Funding.FundingDetails = append(current.Funding.FundingDetails, fundingDetail)
	return c.retry(func() (bool, error) {
		if err := c.projectDao.Update(c.component.BackgroundContext(), false, current); err != nil {
			return true, err
		}
		return false, nil
	})
}

func (c *Coincarp) addNewProject(singleRes ReadDB) error {
	projectDetail := model.Project{
		ProjectInternalInfo: model.ProjectInternalInfo{
			ComponentAutoFillStatus: []string{module},
		},
		Basic: model.ProjectBasic{
			Name:    singleRes.ProjectName,
			LogoURL: singleRes.Logo,
		},
		Funding: model.ProjectFunding{
			FundingDetails: []model.ProjectFundingDetail{
				c.parseFundingDetail(singleRes),
			},
		},
	}
	return c.retry(func() (bool, error) {
		if err := c.projectDao.Add(c.component.BackgroundContext(), false, &projectDetail); err != nil {
			c.component.Logger.WithFields(logrus.Fields{"err": err}).Warn("project add failed")
			return true, err
		}
		return false, nil
	})
}
