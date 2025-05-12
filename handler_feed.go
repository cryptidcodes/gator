package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/cryptidcodes/gatorcli/internal/database"
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
	println("Attempting to fetch feed...")
	println("Creating new HTTP request...")
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	println("Setting header...")
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	println("Executing HTTP request...")
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	println("Reading request data...")
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error: %v", err)
	}
	var result RSSFeed
	println("Unmarshalling...")
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
	println("Adding new feed to database...")
	println("Creating new UUID...")
	newID := uuid.New()

	println("Creating new feed...")
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

	println("Retrieving feed from database...")
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

	println("Retrieving desired feed...")
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	println("Setting feed follow parameters...")
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	}
	println("Creating feed_follows row...")
	row, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	println("Feed followed:")
	println(row.FeedName)
	println(row.UserName)
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
	testFeedURL := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), testFeedURL)
	println(feed.Channel.Title)
	println(feed.Channel.Description)
	for i := 0; i < len(feed.Channel.Item); i++ {
		println(feed.Channel.Item[i].Title)
		println(feed.Channel.Item[i].Description)
	}
	println("Error: ", err)
	return err
}
