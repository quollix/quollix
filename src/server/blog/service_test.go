package blog

import (
	"testing"
	"time"

	"server/configs"
	"server/tools"

	"github.com/quollix/common/assert"
)

var blogServiceTestNow = time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

type blogServiceTestDependencies struct {
	Service     *BlogServiceImpl
	ConfigsRepo *configs.ConfigsRepositoryMock
	Client      *FeedClientMock
	OsWrapper   *tools.CommonOsWrapperMock
}

func setupBlogServiceTestDependencies(t *testing.T) *blogServiceTestDependencies {
	configsRepo := configs.NewConfigsRepositoryMock(t)
	client := NewFeedClientMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	service := &BlogServiceImpl{
		Client:      client,
		ConfigsRepo: configsRepo,
		OsWrapper:   osWrapper,
	}

	return &blogServiceTestDependencies{
		Service:     service,
		ConfigsRepo: configsRepo,
		Client:      client,
		OsWrapper:   osWrapper,
	}
}

func assertBlogServiceTestExpectations(t *testing.T, dependencies *blogServiceTestDependencies) {
	dependencies.ConfigsRepo.AssertExpectations(t)
	dependencies.Client.AssertExpectations(t)
	dependencies.OsWrapper.AssertExpectations(t)
}

func TestBlogService_RefreshLatestArticleDateIfDueSavesDates(t *testing.T) {
	dependencies := setupBlogServiceTestDependencies(t)
	defer assertBlogServiceTestExpectations(t, dependencies)

	articleDate := time.Date(2026, 11, 21, 19, 30, 0, 0, time.UTC)
	dependencies.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.LastBlogFetchAttempt).Return(time.Time{}.Format(time.RFC3339), nil)
	dependencies.OsWrapper.EXPECT().Now().Return(blogServiceTestNow)
	dependencies.ConfigsRepo.EXPECT().SetConfig(configs.ConfigKeys.LastBlogFetchAttempt, blogServiceTestNow.Format(time.RFC3339)).Return(nil)
	dependencies.Client.EXPECT().GetLatestArticleDate().Return(articleDate, nil)
	dependencies.ConfigsRepo.EXPECT().SetConfig(configs.ConfigKeys.LatestBlogArticleDate, "2026-11-21T19:30:00Z").Return(nil)

	err := dependencies.Service.RefreshLatestArticleDateIfDue()
	assert.Nil(t, err)
}

func TestBlogService_RefreshLatestArticleDateIfDueDoesNotFetchWhenRecentAttempt(t *testing.T) {
	dependencies := setupBlogServiceTestDependencies(t)
	defer assertBlogServiceTestExpectations(t, dependencies)

	dependencies.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.LastBlogFetchAttempt).Return(blogServiceTestNow.Add(-time.Hour).Format(time.RFC3339), nil)
	dependencies.OsWrapper.EXPECT().Now().Return(blogServiceTestNow)

	err := dependencies.Service.RefreshLatestArticleDateIfDue()
	assert.Nil(t, err)
}

func TestBlogService_HasUnreadPosts(t *testing.T) {
	lastBlogCheck := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		latestArticleDate string
		expected          bool
	}{
		{"2026-11-21T00:00:00Z", true},
		{"2026-11-20T12:00:00Z", false},
	}

	for _, test := range tests {
		t.Run(test.latestArticleDate, func(t *testing.T) {
			dependencies := setupBlogServiceTestDependencies(t)
			defer assertBlogServiceTestExpectations(t, dependencies)

			dependencies.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.LatestBlogArticleDate).Return(test.latestArticleDate, nil)

			actual, err := dependencies.Service.HasUnreadPosts(lastBlogCheck)

			assert.Nil(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
