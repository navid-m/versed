package feeds

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// Represents a Reddit RSS feed.
type RedditFeed struct {
	Subreddit string
}

// Returns the RSS URL for Reddit.
func (r *RedditFeed) GetFeedURL() string {
	return fmt.Sprintf("https://www.reddit.com/r/%s/.rss", r.Subreddit)
}

// Returns the source name.
func (r *RedditFeed) GetSourceName() string {
	return fmt.Sprintf("Reddit - r/%s", r.Subreddit)
}

// Parses Reddit RSS feed with custom logic for score and comments.
func (r *RedditFeed) ParseFeed(content []byte, sourceID int) ([]FeedItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Reddit feed: %w", err)
	}

	var items []FeedItem
	for _, item := range feed.Items {
		id := generateItemID(item.Link)
		score := extractRedditScore(item.Description)
		commentsCount := extractRedditComments(item.Link)
		title := cleanRedditTitle(item.Title)
		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}
		var curTime = time.Now()
		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   &publishedAt,
			Score:         score,
			CommentsCount: commentsCount,
			CreatedAt:     &curTime,
		}

		if innerLink := extractRedditInnerLink(item.Description); innerLink != "" {
			feedItem.URL = innerLink
		} else if isDirectRedditLink(item.Link) {
			feedItem.URL = fmt.Sprintf("/post/%s", id)
		}

		items = append(items, feedItem)
	}

	return items, nil
}

// Extracts score from Reddit's HTML description.
func extractRedditScore(description string) int {
	re := regexp.MustCompile(`(\d+)\s*points?`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		if score, err := strconv.Atoi(matches[1]); err == nil {
			return score
		}
	}
	return 0
}

// Extracts comment count from Reddit URL.
func extractRedditComments(_ string) int {
	return 0
}

// Removes Reddit-specific formatting from titles.
func cleanRedditTitle(title string) string {
	re := regexp.MustCompile(`^\[.*?\]\s*`)
	return re.ReplaceAllString(title, "")
}

// Extracts the best inner link from Reddit's HTML description.
// Prioritizes i.reddit.com, imgur.com, then other image/video links.
func extractRedditInnerLink(description string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(description))
	if err != nil {
		return ""
	}

	var links []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			links = append(links, href)
		}
	})

	for _, link := range links {
		if strings.Contains(link, "i.reddit.com") {
			return link
		}
	}
	for _, link := range links {
		if strings.Contains(link, "imgur.com") {
			return link
		}
	}
	if len(links) > 0 {
		return links[0]
	}
	return ""
}

// Determines if a URL is a direct Reddit link that should be avoided
func isDirectRedditLink(url string) bool {
	return strings.Contains(url, "reddit.com/r/") && strings.Contains(url, "/comments/")
}

// HN feed structure.
type HackerNewsFeed struct{}

// Get the feed URL for HN.
func (h *HackerNewsFeed) GetFeedURL() string {
	return "https://hnrss.org/frontpage"
}

// Returns the source name.
func (h *HackerNewsFeed) GetSourceName() string {
	return "Hacker News"
}

// Parses the HN RSS feed.
func (h *HackerNewsFeed) ParseFeed(content []byte, sourceID int) ([]FeedItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HN feed: %w", err)
	}

	var items []FeedItem
	for _, item := range feed.Items {
		id := generateItemID(item.Link)

		score := extractHNScore(item.Description)
		commentsCount := extractHNComments(item.Description)

		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}
		var curTime = time.Now()
		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   &publishedAt,
			Score:         score,
			CommentsCount: commentsCount,
			CreatedAt:     &curTime,
		}
		items = append(items, feedItem)
	}

	return items, nil
}

func extractHNScore(description string) int {
	re := regexp.MustCompile(`(\d+)\s*points?`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		if score, err := strconv.Atoi(matches[1]); err == nil {
			return score
		}
	}
	return 0
}

func extractHNComments(description string) int {
	re := regexp.MustCompile(`(\d+)\s*comments?`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		if comments, err := strconv.Atoi(matches[1]); err == nil {
			return comments
		}
	}
	return 0
}

type LobsterFeed struct{}

func (l *LobsterFeed) GetFeedURL() string {
	return "https://lobste.rs/rss"
}

func (l *LobsterFeed) GetSourceName() string {
	return "Lobster.rs"
}

func (l *LobsterFeed) ParseFeed(content []byte, sourceID int) ([]FeedItem, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Lobster feed: %w", err)
	}

	var items []FeedItem
	for _, item := range feed.Items {
		id := generateItemID(item.Link)
		score := extractLobsterScore(item.Description)

		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else {
			publishedAt = time.Now()
		}
		var curTime = time.Now()
		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   &publishedAt,
			Score:         score,
			CommentsCount: 0,
			CreatedAt:     &curTime,
		}
		items = append(items, feedItem)
	}

	return items, nil
}

// Extracts score from Lobster's HTML description.
func extractLobsterScore(description string) int {
	re := regexp.MustCompile(`(\d+)\s*points?`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		if score, err := strconv.Atoi(matches[1]); err == nil {
			return score
		}
	}
	return 0
}

// Manages multiple feed sources.
type FeedManager struct {
	Sources []FeedSourceInterface
}

// Defines the interface for feed sources.
type FeedSourceInterface interface {
	GetFeedURL() string
	GetSourceName() string
	ParseFeed(content []byte, sourceID int) ([]FeedItem, error)
}

// GenericRSSFeed represents a generic RSS feed from any URL
type GenericRSSFeed struct {
	URL  string
	Name string
}

// GetFeedURL returns the RSS URL for the generic feed
func (g *GenericRSSFeed) GetFeedURL() string {
	return g.URL
}

// GetSourceName returns the source name for the generic feed
func (g *GenericRSSFeed) GetSourceName() string {
	return g.Name
}

// ParseFeed parses the RSS feed using the generic parser
func (g *GenericRSSFeed) ParseFeed(content []byte, sourceID int) ([]FeedItem, error) {
	return ParseFeedWithParser(content, sourceID, g.Name)
}

// Creates a new feed manager with default sources.
func NewFeedManager() *FeedManager {
	return &FeedManager{
		Sources: []FeedSourceInterface{
			&RedditFeed{Subreddit: "programming"},
			&RedditFeed{Subreddit: "technology"},
			&HackerNewsFeed{},
			&LobsterFeed{},
		},
	}
}

// CreateRedditFeed creates a new Reddit feed for a given subreddit
func CreateRedditFeed(subreddit string) FeedSourceInterface {
	return &RedditFeed{Subreddit: subreddit}
}

// CreateGenericRSSFeed creates a generic RSS feed from a URL
func CreateGenericRSSFeed(url, name string) *GenericRSSFeed {
	return &GenericRSSFeed{
		URL:  url,
		Name: name,
	}
}
