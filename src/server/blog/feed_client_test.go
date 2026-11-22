package blog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestRSSClient_GetLatestArticleDateReturnsNewestPublicationDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, `<rss version="2.0"><channel><item><pubDate>Sat, 21 Nov 2026 00:00:00 +0000</pubDate></item><item><pubDate>Mon, 23 Nov 2026 00:00:00 +0000</pubDate></item></channel></rss>`)
	}))
	defer server.Close()

	client := NewRSSClient(server.Client(), server.URL)
	actual, err := client.GetLatestArticleDate()

	assert.Equal(t, nil, err)
	assert.Equal(t, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC), actual)
}

func TestMockFeedClient_ReturnsDateFrom24HoursAgo(t *testing.T) {
	client := &MockFeedClient{}
	date, err := client.GetLatestArticleDate()
	assert.Nil(t, err)
	assert.True(t, date.Before(time.Now().UTC()))
	assert.True(t, time.Since(date) >= 24*time.Hour)
	assert.True(t, time.Since(date) < 24*time.Hour+time.Second)
}
