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

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
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

		//log.Println("Found post", item.Title, "on feed", feed.Name)
	}
	log.Printf("Feed %s collected, %v posts found, %v posts updated", feed.Name, len(rssFeed.Channel.Item), updatedPosts)
	return updatedPosts, nil
}
