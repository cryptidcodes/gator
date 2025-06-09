package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cryptidcodes/gator/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	var result RSSFeed
	err = xml.Unmarshal(data, &result)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	result.Channel.Title = html.UnescapeString(result.Channel.Title)
	result.Channel.Description = html.UnescapeString(result.Channel.Description)
	for i := 0; i < len(result.Channel.Item); i++ {
		result.Channel.Item[i].Title = html.UnescapeString(result.Channel.Item[i].Title)
		result.Channel.Item[i].Description = html.UnescapeString(result.Channel.Item[i].Description)
	}

	return &result, nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		log.Fatal("syntax: addfeed requires 2 args")
	}
	newID := uuid.New()

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        newID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return err
	}

	feed, err = s.db.GetFeedByID(context.Background(), newID)

	if err != nil {
		return err
	}
	fmt.Printf("New feed created!\n")
	fmt.Printf("UUID: %v\n", feed.ID)
	fmt.Printf("CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf("UpdatedAt: %v\n", feed.UpdatedAt)
	fmt.Printf("Name: %v\n", feed.Name)
	fmt.Printf("Url: %v\n", feed.Url)
	fmt.Printf("UserID: %v\n", feed.UserID)

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	}
	s.db.CreateFeedFollow(context.Background(), params)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.Args) != 0 {
		log.Fatal("syntax: feeds does not accept args")
	}
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for i := 0; i < len(feeds); i++ {
		user, err := s.db.GetUserByID(context.Background(), feeds[i].UserID)
		if err != nil {
			return err
		}
		println(feeds[i].Name)
		println(feeds[i].Url)
		println("Created by: ", user.Name)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		log.Fatal("syntax: follow requires 1 arg (url)")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	}
	row, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%v followed: %v\n", row.UserName, row.FeedName)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	// deletes a feed follow for the currently logged in user

	// ensure we were passed exactly 1 arg (url)
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v {feedURL}", cmd.Name)
	}

	// retrieve the feed ID
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	// set the command parameters
	params := database.UnfollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	}
	err = s.db.Unfollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	rows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for i := 0; i < len(rows); i++ {
		println(rows[i].FeedName)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v {time_between_reqs}: duration string", cmd.Name)
	}
	// parse the time arg
	dur, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(dur)
	for ; ; <-ticker.C {
		println("Aggin...")
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		log.Println(err)
	}
	for i := 0; i < len(feeds); i++ {
		// determine the next feed to fetch
		dbFeed, err := s.db.GetNextFeedToFetch(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		// fetch current state of feed
		fmt.Printf("Fetching from %v\n", dbFeed.Name)
		rssFeed, err := fetchFeed(context.Background(), dbFeed.Url)
		if err != nil {
			log.Println(err)
		}

		// mark the feed as fetched
		s.db.MarkFeedFetched(context.Background(), dbFeed.ID)

		// create posts table entries for any posts that dont have entries already
		for i := 0; i < len(rssFeed.Channel.Item); i++ {
			_, err := s.db.GetPostByUrl(context.Background(), rssFeed.Channel.Item[i].Link)
			if err != nil {
				params := database.CreatePostParams{
					ID:          uuid.New(),
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
					Title:       rssFeed.Channel.Item[i].Title,
					Url:         rssFeed.Channel.Item[i].Link,
					Description: rssFeed.Channel.Item[i].Description,
					PublishedAt: rssFeed.Channel.Item[i].PubDate,
					FeedID:      dbFeed.ID,
				}
				_, err := s.db.CreatePost(context.Background(), params)
				if err != nil {
					println(err)
				}
				println(params.Title)
				println(params.PublishedAt)
				println(params.Url)
			}
		}
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	// prints posts using GetPostsForUser

	// ensure maximum of 1 arg was passed
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: %v {num_posts}", cmd.Name)
	}

	// retrieve current user's ID
	user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	// create GetPostsForUser params
	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  2,
	}

	// if a limit arg was passed, set the limit parameter to match
	if len(cmd.Args) == 1 {
		i, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return err
		}
		params.Limit = int32(i)
	}

	// get the posts
	row, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}
	for i := 0; i < len(row); i++ {
		println(row[i].Name)
		println(row[i].Title)
		println(row[i].PublishedAt)
		println(row[i].Url)
		// println(row[i].Description)
	}
	return nil
}