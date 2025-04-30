package main

import (
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	// if the commands args slice length is not 1, return an error
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)

	}
	username := cmd.Args[0]
	err := s.cfg.SetUser(username)

	if err != nil {
		return fmt.Errorf("couldn't login user: %v", err)
	}
	fmt.Printf("User: %v has logged in!\n", username)
	return nil
}
