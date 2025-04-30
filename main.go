package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cryptidcodes/gatorcli/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		println(err)
	}
	// create a new state
	s := state{
		cfg: &cfg,
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
		os.Exit(1)
	}

	// confirm the command the user is trying to run exists
	_, exists := cmds.cmdmap[os.Args[1]]
	if !exists {
		log.Fatal("unregistered command, please use a registered command\n")
	}

	// build the func
	userCmd := command{
		Name: os.Args[1],
		Args: make([]string, 0),
	}

	// add args
	if len(os.Args) > 2 {
		userCmd.Args = os.Args[2:]
	}
	// run the func
	err = cmds.run(&s, userCmd)
	if err != nil {
		log.Fatal(err)
	}
}
