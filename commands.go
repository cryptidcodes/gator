package main

import (
	"context"
	"errors"

	"github.com/cryptidcodes/gatorcli/internal/database"
)

type command struct {
	Name string
	Args []string
}

type commands struct {
	cmdmap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	// runs a given command with the provided state if it exists
	handlerFunc, exists := c.cmdmap[cmd.Name]
	if !exists {
		return errors.New("command does not exist")
	}
	return handlerFunc(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	// registers a new handler function for a command name
	c.cmdmap[name] = f
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		// check if a user is logged in
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}
