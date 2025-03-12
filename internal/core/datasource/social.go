package datasource

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wyt-labs/wyt-core/internal/core/dao"
	"github.com/wyt-labs/wyt-core/internal/pkg/base"
	"github.com/wyt-labs/wyt-core/pkg/util"
)

const (
	invalidUrl    = "invalid url"
	invalidCode   = "404"
	userSuffix    = "https://api.github.com/users"
	repoSuffix    = "https://github.com"
	githubCacheID = "github_cache"
)

type ProjectGhCache struct {
	UpdateTime int64
	// only cacheDao view info
	ProjectGhInfoMap map[string]*GhInfo
}

type ghInfoCacheWrapper struct {
	id          string
	validGhInfo *GhInfo
}

type Social struct {
	baseComponent      *base.Component
	lock               *sync.RWMutex
	projectDao         *dao.ProjectDao
	cacheDao           *dao.SystemCacheDao
	goPool             *util.GoPool
	socialInfoPool     *sync.Pool
	ghInfoCacheCh      chan *ghInfoCacheWrapper
	oldCache           ProjectGhCache
	updateProjectCount int64
	// id -> url
	projectUrlMap map[string]*ProjectUrl

	// id -> github info
	ghInfoMap map[string]*GhInfo
}

func (s *Social) GetGithubInfo(id string) *GhInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ghInfoMap[id]
}

func NewSocial(baseComponent *base.Component, projectDao *dao.ProjectDao, systemCache *dao.SystemCacheDao) (*Social, error) {
	pool := &sync.Pool{
		New: func() interface{} {
			return new(GhInfo)
		},
	}
	s := &Social{
		baseComponent:  baseComponent,
		lock:           new(sync.RWMutex),
		projectUrlMap:  make(map[string]*ProjectUrl),
		ghInfoMap:      make(map[string]*GhInfo),
		ghInfoCacheCh:  make(chan *ghInfoCacheWrapper, 1000),
		projectDao:     projectDao,
		goPool:         util.NewGoPool(baseComponent.Config.Datasource.Social.Github.ConcurrencyLimit),
		socialInfoPool: pool,
		cacheDao:       systemCache,
	}
	baseComponent.RegisterLifecycleHook(s)
	return s, nil
}

func (s *Social) Start() error {
	if s.baseComponent.Config.Datasource.Social.Disable {
		return nil
	}

	s.baseComponent.SafeGo(func() {
		if err := s.loadAllProjectUrl(); err != nil {
			s.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to load all project url")
			s.baseComponent.ComponentShutdown()
		}

		// load cache from db
		var ghCache ProjectGhCache
		if err := s.cacheDao.Get(s.baseComponent.BackgroundContext(), githubCacheID, &ghCache); err != nil {
			if err != mongo.ErrNoDocuments {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to load github cache")
				s.baseComponent.ComponentShutdown()
			}
		}
		s.oldCache = ghCache
		refreshInternal := time.Duration(s.baseComponent.Config.Datasource.Social.Github.RefreshInterval)
		now := time.Now()
		if s.oldCache.UpdateTime+int64(refreshInternal.Seconds()) < now.Unix() {
			// fetch github data
			if err := s.fetchGithubData(); err != nil {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to fetch github data")
				s.baseComponent.ComponentShutdown()
			}
			s.baseComponent.Logger.WithFields(logrus.Fields{
				"project count": len(s.ghInfoMap),
				"time cost":     time.Since(now)}).Info("fetch github data success")
		}

		_, err := s.baseComponent.Cron.AddFunc(s.baseComponent.Config.Datasource.Social.Github.RefreshCron, func() {
			if err := s.fetchGithubData(); err != nil {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"err": err,
				}).Error("Failed to fetch active user data")
			}
		})
		if err != nil {
			s.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to add ActiveUserDataRefreshCron task")
			s.baseComponent.ComponentShutdown()
			return
		}
	})

	// start listen gh info cache
	s.baseComponent.SafeGo(func() {
		if err := s.listenGhInfoCache(); err != nil {
			s.baseComponent.Logger.WithFields(logrus.Fields{
				"err": err,
			}).Error("Failed to listen github info cache")
			s.baseComponent.ComponentShutdown()
		}
	})
	return nil
}

func (s *Social) listenGhInfoCache() error {
	curUpdateSize := 0
	newCacheM := make(map[string]*GhInfo)

	// deep copy old cache to new cache
	s.lock.RLock()
	for k, v := range s.oldCache.ProjectGhInfoMap {
		newCacheM[k] = v
	}
	s.lock.RUnlock()

	ticker := time.NewTicker(time.Duration(s.baseComponent.Config.Datasource.Social.Github.CacheTimeout))
	defer ticker.Stop()
	for {
		select {
		case <-s.baseComponent.Ctx.Done():
			return nil
		case ghInfoCache := <-s.ghInfoCacheCh:
			curUpdateSize++
			newCacheM[ghInfoCache.id] = ghInfoCache.validGhInfo
			if curUpdateSize >= s.baseComponent.Config.Datasource.Social.Github.Checkpoint {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"curUpdateSize": curUpdateSize,
				}).Info("update github cache to DB")

				// put new cache to cacheDao
				newCache := &ProjectGhCache{
					UpdateTime:       time.Now().Unix(),
					ProjectGhInfoMap: newCacheM,
				}
				if err := s.cacheDao.Put(s.baseComponent.BackgroundContext(), githubCacheID, newCache); err != nil {
					return fmt.Errorf("put new cache to cache DB failed, err is %s", err)
				}
				curUpdateSize = 0
			}
		case <-ticker.C:
			if curUpdateSize != 0 {
				newCache := &ProjectGhCache{
					UpdateTime:       time.Now().Unix(),
					ProjectGhInfoMap: newCacheM,
				}
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"curUpdateSize": curUpdateSize,
				}).Info("update github cache to DB")

				if err := s.cacheDao.Put(s.baseComponent.BackgroundContext(), githubCacheID, newCache); err != nil {
					return fmt.Errorf("put new cache to cache DB failed, err is %s", err)
				}
				curUpdateSize = 0
			}
		}
	}
}

func (s *Social) Stop() error {
	return nil
}

type ProjectUrl struct {
	name       string
	githubUrl  string
	twitterUrl string
}

type ProjectBasicInfo struct {
	Name string `json:"name" bson:"name"`
}

type ProjectsUrlsElement struct {
	ID           string              `json:"id" bson:"_id,omitempty"`
	Basic        ProjectBasicInfo    `json:"basic" bson:"basic"`
	RelatedLinks ProjectRelatedLinks `json:"related_links" bson:"related_links"`
}

type ProjectRelatedLinks []LinkInfo

type LinkInfo struct {
	Type string `json:"type" bson:"type"`
	Link string `json:"link" bson:"link"`
}

// MetricsInfo todo: add more metrics info
type MetricsInfo struct {
	ActiveDevelopers      int64
	AveTxFeeseLast24Hours float64 //average transaction fee last 24 hours(usd)
	VolumeLast24Hours     float64 //real volume last 24 hours(usd)
	MarketCap             float64
	Ath                   float64 // All Time high(usd)
	Cl                    float64 // Cycle Low(usd)
	UniqueAddress         int64   // project all address count
}

type GhInfo struct {
	InvalidUrl         bool
	NeedRetry          bool
	Name               string
	GithubCommits      int64
	GithubStars        int64
	GithubForks        int64
	GithubContributors int64
	GithubFollowers    int64
}

// TwitterInfo todo: add twitter info
type TwitterInfo struct {
	IsValidUrl       bool
	NeedRetry        bool
	Name             string
	TwitterFollowers int64
}

// RedditInfo todo: add reddit info
type RedditInfo struct {
	RedditMembers int64
}

func (s *Social) loadAllProjectUrl() error {
	var list []*ProjectsUrlsElement

	_, err := s.projectDao.CustomList(s.baseComponent.BackgroundContext(), false, 0, 0, nil, nil, &list)

	if err != nil {
		return err
	}

	for _, e := range list {
		for _, link := range e.RelatedLinks {
			if link.Type == "GitHub" {
				if _, ok := s.projectUrlMap[e.ID]; !ok {
					s.projectUrlMap[e.ID] = &ProjectUrl{
						name:      e.Basic.Name,
						githubUrl: link.Link,
					}
				} else {
					s.projectUrlMap[e.ID].githubUrl = link.Link
				}
				// init github info
				s.ghInfoMap[e.ID] = &GhInfo{}
			} else if link.Type == "Twitter" {
				if _, ok := s.projectUrlMap[e.ID]; !ok {
					s.projectUrlMap[e.ID] = &ProjectUrl{
						name:       e.Basic.Name,
						twitterUrl: link.Link,
					}
				} else {
					s.projectUrlMap[e.ID].twitterUrl = link.Link
				}
			} else {
				s.projectUrlMap[e.ID] = &ProjectUrl{}
			}
		}
	}
	s.baseComponent.Logger.Infof("load %d project url success", len(s.projectUrlMap))
	return nil
}

func (s *Social) getUpdateGhProject() (map[string]*ProjectUrl, error) {
	for id, ghInfo := range s.oldCache.ProjectGhInfoMap {
		if ghInfo.NeedRetry || ghInfo.InvalidUrl {
			continue
		}
		// load github info from db
		s.ghInfoMap[id] = ghInfo
	}

	needUpdateM := make(map[string]*ProjectUrl)
	// filter need update project
	for id, url := range s.projectUrlMap {
		if url.githubUrl == "" {
			continue
		}
		if _, ok := s.oldCache.ProjectGhInfoMap[id]; !ok {
			needUpdateM[id] = url
		} else {
			if s.oldCache.ProjectGhInfoMap[id].NeedRetry {
				needUpdateM[id] = url
			}
		}
	}
	return needUpdateM, nil
}

func (s *Social) fetchGithubData() error {
	now := time.Now()
	// load all update project url
	needUpdateM, err := s.getUpdateGhProject()
	if err != nil {
		return err
	}
	s.baseComponent.Logger.Infof("start update %d project", len(needUpdateM))

	for id, url := range needUpdateM {
		s.goPool.Add()
		go func(id string, url *ProjectUrl) {
			defer s.goPool.Done()
			info := s.socialInfoPool.Get().(*GhInfo)
			info.Name = url.name
			defer func() {
				// put info to sync pool
				info = &GhInfo{}
				s.socialInfoPool.Put(info)
			}()
			if url.githubUrl != "" {
				// fill github info
				err := s.parseGithub(url.githubUrl, info)
				if err != nil {
					s.baseComponent.Logger.Warnf("parse github url %s failed, err is %s", url.githubUrl, err)
					// if url is invalid, it is certain that the user's URL is invalid,
					// delete key in ghInfoMap
					if strings.Contains(err.Error(), invalidUrl) {
						info = &GhInfo{Name: url.name, InvalidUrl: true}
					} else {
						info.NeedRetry = true
					}
				}
			} else {
				info = &GhInfo{Name: url.name, InvalidUrl: true}
			}

			s.lock.Lock()
			s.ghInfoMap[id] = info
			s.lock.Unlock()

			s.ghInfoCacheCh <- &ghInfoCacheWrapper{
				id:          id,
				validGhInfo: info,
			}

			atomic.AddInt64(&s.updateProjectCount, 1)

			if url.githubUrl != "" {
				s.baseComponent.Logger.WithFields(logrus.Fields{
					"project":             url.githubUrl,
					"github followers":    info.GithubFollowers,
					"github stars":        info.GithubStars,
					"github forks":        info.GithubForks,
					"github commits":      info.GithubCommits,
					"github contributors": info.GithubContributors,
				}).Debugf("fetch github data success")
			}
		}(id, url)
	}
	s.goPool.Wait()

	s.baseComponent.Logger.WithFields(logrus.Fields{
		"valid project count": s.updateProjectCount,
		"time cost":           time.Since(now)}).
		Info("fetch all projects github data success")

	return nil
}

// convert string to int64 in us number format
func parseCount(str string) (int64, error) {
	if str == "" {
		return 0, fmt.Errorf("input element is empty")
	}
	var (
		number int64
		err    error
	)
	// check if the last character is a letter
	if !unicode.IsDigit([]rune(str)[len(str)-1]) {
		number, err = formatNumber(str)
		if err != nil {
			return 0, err
		}
		return number, nil
	}

	str = strings.ReplaceAll(str, ",", "")

	number, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func formatNumber(value string) (int64, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	suffixes := map[string]int64{
		"k": 1000,          // kilo
		"m": 1000000,       // million
		"b": 1000000000,    // billion
		"t": 1000000000000, // trillion
	}

	for suffix, multiplier := range suffixes {
		if strings.HasSuffix(value, suffix) {
			// Remove the suffix from the value
			value = strings.TrimSuffix(value, suffix)

			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, err
			}
			// Multiply the numeric part with the corresponding multiplier
			res := num * float64(multiplier)
			return int64(res), nil
		}
	}

	// If no suffix was found, return an error
	return 0, fmt.Errorf("invalid suffix %s", value[len(value)-1:])
}

func (s *Social) parseGithub(url string, info *GhInfo) error {
	var (
		err               error
		starsCount        int64
		forksCount        int64
		commitsCount      int64
		contributorsCount int64
		followersCount    int64
	)
	user, repo, err := s.formatUrl(url)
	if err != nil {
		return err
	}

	if user != "" {
		userUrl := fmt.Sprintf("%s/%s", userSuffix, user)
		followersCount, err = s.parseGithubUserInfo(userUrl)
		if err != nil {
			return err
		}
		info.GithubFollowers = followersCount
	}

	if repo != "" {
		repoUrl := fmt.Sprintf("%s/%s/%s", repoSuffix, user, repo)
		starsCount, forksCount, commitsCount, contributorsCount, err = s.parseGithubRepoInfo(repoUrl)
		if err != nil {
			return err
		}
		info.GithubStars = starsCount
		info.GithubForks = forksCount
		info.GithubCommits = commitsCount
		info.GithubContributors = contributorsCount
	}

	return nil
}

func (s *Social) parseGithubRepoInfo(repoUrl string) (int64, int64, int64, int64, error) {
	var (
		starsElement        string
		forksElement        string
		commitsElement      string
		contributorsElement string
	)

	err := util.Retry(2*time.Second, s.baseComponent.Config.Datasource.Social.Github.RetryLimit, func() (bool, error) {
		// 1. get html body
		resp, err := s.getHtmlResponse(repoUrl)
		if err != nil {
			// if url is invalid, need not retry
			if strings.Contains(err.Error(), invalidCode) {
				return false, fmt.Errorf("%s: %s", invalidUrl, repoUrl)
			}
			return true, err
		}
		doc, err := htmlquery.Parse(strings.NewReader(resp.String()))
		if err != nil {
			return true, fmt.Errorf("parse html failed:%s", err)
		}

		// 2. get stars, forks, commits, contributors
		sel := htmlquery.FindOne(doc, `//*[@id="repo-stars-counter-star"]`)
		if sel == nil {
			return false, fmt.Errorf("find stars failed element failed")
		}
		starsElement = htmlquery.SelectAttr(sel, "title")
		sel = htmlquery.FindOne(doc, `//*[@id="repo-network-counter"]`)
		if sel == nil {
			return false, fmt.Errorf("find forks element failed")
		}
		forksElement = htmlquery.SelectAttr(sel, "title")

		sel = htmlquery.FindOne(doc, "//a[contains(@href, '/commits')]/span/strong")
		if sel == nil {
			return false, fmt.Errorf("find commits element failed")
		}
		commitsElement = htmlquery.InnerText(sel)

		// contributors may not exist, so ignore error
		sel = htmlquery.FindOne(doc, "//a[contains(@href, '/graphs/contributors')]/span")
		if sel != nil {
			contributorsElement = htmlquery.InnerText(sel)
		}
		return false, nil
	})

	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Failed to get repo page: %v\n", err)
	}
	starsCount, err := parseCount(starsElement)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Failed to parse stars count: %v\n", err)
	}
	forksCount, err := parseCount(forksElement)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Failed to parse forks count: %v\n", err)
	}
	commitsCount, err := parseCount(commitsElement)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("Failed to parse commits count: %v\n", err)
	}
	contributorsCount := int64(0)
	if contributorsElement != "" {
		contributorsCount, err = parseCount(contributorsElement)
		if err != nil {
			return 0, 0, 0, 0, fmt.Errorf("Failed to parse contributors count: %v\n", err)
		}
	}

	return starsCount, forksCount, commitsCount, contributorsCount, nil
}

type User struct {
	Followers int `json:"followers"`
}

func (s *Social) parseGithubUserInfo(userUrl string) (int64, error) {
	//var followersElement string
	var user User
	err := util.Retry(2*time.Second, s.baseComponent.Config.Datasource.Social.Github.RetryLimit, func() (bool, error) {
		// 1. get html body
		resp, err := s.getHtmlResponse(userUrl)
		if err != nil {
			// if url is invalid, need not retry
			if strings.Contains(err.Error(), invalidCode) {
				return false, fmt.Errorf("%s: %s", invalidUrl, userUrl)
			}
			return true, err
		}

		err = json.Unmarshal(resp.Body(), &user)
		if err != nil {
			return true, err
		}
		return false, nil
	})

	if err != nil {
		return 0, fmt.Errorf("Failed to get user page: %v\n", err)
	}

	return int64(user.Followers), nil
}

func (s *Social) getHtmlResponse(url string) (*resty.Response, error) {
	httpClient := resty.New()
	httpClient.SetTimeout(time.Duration(s.baseComponent.Config.Datasource.Social.Github.HttpTimeout))

	// set auth token randomly
	ghTokens := s.baseComponent.Config.Datasource.Social.Github.APIKeys
	rand.Seed(time.Now().UnixNano())
	tokenIndex := rand.Intn(len(ghTokens))
	httpClient.SetAuthToken(ghTokens[tokenIndex])

	resp, err := httpClient.R().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("http request failed, code: %d", resp.StatusCode())
	}

	return resp, nil
}

func (s *Social) formatUrl(url string) (string, string, error) {
	// get user url
	url = strings.TrimSpace(url)
	re := regexp.MustCompile(`^(https?://(?:www\.)?github.com/([^/]+)(?:/([^/]+))?)`)

	matches := re.FindStringSubmatch(url)
	if len(matches) != 4 {
		return "", "", fmt.Errorf("invalid url: %s", url)
	}
	user := matches[2]
	repo := matches[3]
	return user, repo, nil
}
