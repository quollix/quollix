package blog

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const RSSFeedURL = "https://quollix.org/blog/index.xml"

type FeedClient interface {
	GetLatestArticleDate() (time.Time, error)
}

type RSSClientImpl struct {
	HttpClient *http.Client
	feedURL    string
}

func NewRSSClient(httpClient *http.Client, feedURL string) *RSSClientImpl {
	return &RSSClientImpl{HttpClient: httpClient, feedURL: feedURL}
}

type rssFeed struct {
	Items []rssItem `xml:"channel>item"`
}

type rssItem struct {
	PubDate string `xml:"pubDate"`
}

func (c *RSSClientImpl) GetLatestArticleDate() (time.Time, error) {
	response, err := c.HttpClient.Get(c.feedURL)
	if err != nil {
		return time.Time{}, u.Logger.NewError(err.Error())
	}
	defer u.Close(response.Body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return time.Time{}, u.Logger.NewError("RSS feed request returned an unsuccessful status", "status", response.Status)
	}

	return parseRSSLatestArticleDate(response.Body)
}

func parseRSSLatestArticleDate(reader io.Reader) (time.Time, error) {
	var feed rssFeed
	if err := xml.NewDecoder(reader).Decode(&feed); err != nil {
		return time.Time{}, u.Logger.NewError(err.Error())
	}
	if len(feed.Items) == 0 {
		return time.Time{}, u.Logger.NewError("RSS feed contains no articles")
	}

	var latest time.Time
	for _, item := range feed.Items {
		publishedAt, err := parsePublicationDate(item.PubDate)
		if err != nil {
			return time.Time{}, err
		}
		if publishedAt.After(latest) {
			latest = publishedAt
		}
	}
	return latest.UTC(), nil
}

func parsePublicationDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, u.Logger.NewError("invalid RSS publication date", "date", value)
}

type MockFeedClient struct{}

func (c *MockFeedClient) GetLatestArticleDate() (time.Time, error) {
	return time.Now().UTC().Add(-24 * time.Hour), nil
}
