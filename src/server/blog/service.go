package blog

import (
	"time"

	u "github.com/quollix/common/utils"
	"server/configs"
)

type BlogService interface {
	RefreshLatestArticleDateIfDue() error
	HasUnreadPosts(lastBlogCheck time.Time) (bool, error)
}

const feedRefreshInterval = 6 * time.Hour

type BlogServiceImpl struct {
	Client      FeedClient
	ConfigsRepo configs.ConfigsRepository
	OsWrapper   u.OsWrapper
}

func (s *BlogServiceImpl) RefreshLatestArticleDateIfDue() error {
	lastFetchAttemptValue, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.LastBlogFetchAttempt)
	if err != nil {
		return err
	}
	lastFetchAttempt, err := time.Parse(time.RFC3339, lastFetchAttemptValue)
	if err != nil {
		return err
	}
	now := s.OsWrapper.Now().UTC()
	if now.Sub(lastFetchAttempt) < feedRefreshInterval {
		return nil
	}

	latestArticleDate, err := s.Client.GetLatestArticleDate()
	if err != nil {
		return err
	}
	if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.LastBlogFetchAttempt, now.Format(time.RFC3339)); err != nil {
		return err
	}
	return s.ConfigsRepo.SetConfig(configs.ConfigKeys.LatestBlogArticleDate, latestArticleDate.UTC().Format(time.RFC3339))
}

func (s *BlogServiceImpl) HasUnreadPosts(lastBlogCheck time.Time) (bool, error) {
	latestArticleDateValue, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.LatestBlogArticleDate)
	if err != nil {
		return false, err
	}
	latestArticleDate, err := time.Parse(time.RFC3339, latestArticleDateValue)
	if err != nil {
		return false, err
	}

	return latestArticleDate.After(lastBlogCheck), nil
}
