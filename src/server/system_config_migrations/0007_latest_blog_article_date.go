package system_config_migrations

import (
	"time"

	"server/configs"

	u "github.com/quollix/common/utils"
)

func (s *SystemConfigMigrationsProviderImpl) createLatestBlogArticleDate() error {
	u.Logger.Info(systemConfigMigrationLogMessage, "step", 7, "action", "create blog feed configs")
	// Keep the DEV startup snapshot's blog date slightly ahead of zero so reset tests start with unread posts.
	initialArticleDate := time.Time{}.UTC().Add(time.Hour)
	if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.LatestBlogArticleDate, initialArticleDate.Format(time.RFC3339)); err != nil {
		return err
	}
	return s.ConfigsRepo.SetConfig(configs.ConfigKeys.LastBlogFetchAttempt, time.Time{}.UTC().Format(time.RFC3339))
}
