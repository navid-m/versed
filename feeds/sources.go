package feeds

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

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

		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   publishedAt,
			Score:         score,
			CommentsCount: commentsCount,
			CreatedAt:     time.Now(),
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

		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   publishedAt,
			Score:         score,
			CommentsCount: commentsCount,
			CreatedAt:     time.Now(),
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

		feedItem := FeedItem{
			ID:            id,
			SourceID:      sourceID,
			Title:         item.Title,
			URL:           item.Link,
			Description:   item.Description,
			Author:        item.Author.Name,
			PublishedAt:   publishedAt,
			Score:         score,
			CommentsCount: 0,
			CreatedAt:     time.Now(),
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
