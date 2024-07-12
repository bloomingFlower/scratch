package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bloomingFlower/rssagg/internal/database"
	"github.com/google/uuid"
)

type RunScrapingRequest struct {
	Concurrency int32
}

type RunScrapingResponse struct {
	Status string
}

func (s *server) runScraping(ctx context.Context, req *RunScrapingRequest) (*RunScrapingResponse, error) {
	go startScraping(s.DB, int(req.Concurrency), 0)
	return &RunScrapingResponse{Status: "Scraping started"}, nil
}

func startScraping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
) {
	log.Printf("Scraping on %v go routines every %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Println("error fetching feeds:", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			feed := feed
			go func() {
				updatedPosts, err := scrapeFeed(db, wg, feed)
				if err != nil {
					log.Printf("Error scraping feed: %v", err)
					return
				}
				log.Printf("Updated %d posts", updatedPosts)
			}()
		}
		wg.Wait()

		if timeBetweenRequest == 0 {
			break
		}
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) (int, error) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("Error marking feed as fetched:", err)
		return 0, err
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("Error fetching feed:", err)
		return 0, err
	}

	updatedPosts := 0
	for _, item := range rssFeed.Channel.Item {
		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("couldn't parse date %v with err %v\n", item.PubDate, err)
			continue
		}

		// Create a new post in the database
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),       // Generate a new UUID for the post ID
			CreatedAt:   time.Now().UTC(), // Set the creation time to the current time in UTC
			UpdatedAt:   time.Now().UTC(), // Set the update time to the current time in UTC
			Title:       item.Title,       // Set the post title to the item's title
			Description: description,
			PublishedAt: pubAt,
			Url:         item.Link,
			FeedID:      feed.ID,
		})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Println("failed to create post: ", err)
		} else {
			updatedPosts++
		}
	}

	if updatedPosts > 0 {
		log.Printf("Feed %s collected, %v new posts found", feed.Name, updatedPosts)
	}

	return updatedPosts, nil
}
