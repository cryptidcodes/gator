package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cryptidcodes/gatorcli/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	// if the commands args slice length is not 1, return an error
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)

	}
	user, err := s.db.GetUserByName(context.Background(), cmd.Args[0])
	if err != nil {
		log.Fatal("user does not exist!")
	}

	err = s.cfg.SetUser(user.Name)

	if err != nil {
		return fmt.Errorf("couldn't login user: %v", err)
	}
	fmt.Printf("User: %v has logged in!\n", user.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	// ensure a single name arg was passed
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	// check if user is already registered
	user, _ := s.db.GetUserByName(context.Background(), cmd.Args[0])
	if user.Name == cmd.Args[0] {
		log.Fatal("User already exists!")
	}

	// set user parameters
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	// create the user database entry
	s.db.CreateUser(context.Background(), params)

	// set the user as currently logged in
	s.cfg.SetUser(cmd.Args[0])

	fmt.Println("New user registered!")
	fmt.Printf("UUID: %v\n", params.ID)
	fmt.Printf("CreatedAt: %v\n", params.CreatedAt)
	fmt.Printf("UpdatedAt: %v\n", params.UpdatedAt)
	fmt.Printf("Name: %v\n", params.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	println("users table reset successfully")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(users); i++ {
		username := users[i].Name
		if username == s.cfg.CurrentUserName {
			println(username + " (current)")
			continue
		}
		println(username)
	}
	return nil
}
