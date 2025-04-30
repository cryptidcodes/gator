package main

import (
	"fmt"
	"os"

	"github.com/cryptidcodes/gatorcli/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmdmap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	// runs a given command with the provided state if it exists
	handlerFunc, exists := c.cmdmap[cmd.name]
	if !exists {
		return fmt.Errorf("command does not exist")
	}
	return handlerFunc(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	// registers a new handler function for a command name
	c.cmdmap[name] = f
}

func handlerLogin(s *state, cmd command) error {
	// if the commands args slice length is not 1, return an error
	if len(cmd.args) != 1 {
		return fmt.Errorf("login command expects 1 arg: username")
	}
	username := cmd.args[0]
	err := s.config.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v has logged in!\n", username)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		println(err)
	}
	// create a new state
	s := state{
		config: &cfg,
	}

	// create an instance of the commands struct and initialize the cmdmap
	cmds := commands{
		cmdmap: make(map[string]func(*state, command) error),
	}

	// register commands
	cmds.register("login", handlerLogin)

	// confirm the user input at least two args. Example: gator login
	if len(os.Args) < 2 {
		fmt.Printf("you must input a command to use gator\n")
		return
	}

	// confirm the command the user is trying to run exists
	_, exists := cmds.cmdmap[os.Args[1]]
	if !exists {
		fmt.Printf("unregistered command, please use a registered command\n")
		return
	}

	// build the func
	userCmd := command{
		name: os.Args[1],
		args: make([]string, 0),
	}

	// add args
	if len(os.Args) > 2 {
		userCmd.args = os.Args[2:]
	}
	// run the func
	err = cmds.run(&s, userCmd)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
