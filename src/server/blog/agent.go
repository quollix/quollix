package blog

import (
	"time"

	u "github.com/quollix/common/utils"
)

const agentCheckInterval = 5 * time.Minute

type Agent interface {
	StartBlogAgentDaemon()
}

type AgentImpl struct {
	BlogService BlogService
	OsWrapper   u.OsWrapper
}

func (a *AgentImpl) StartBlogAgentDaemon() {
	go a.run()
}

func (a *AgentImpl) run() {
	for {
		if err := a.BlogService.RefreshLatestArticleDateIfDue(); err != nil {
			u.Logger.Error(err)
		}
		a.OsWrapper.Sleep(agentCheckInterval)
	}
}
