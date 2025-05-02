package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/cryptidcodes/gatorcli/internal/config"
	"github.com/cryptidcodes/gatorcli/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		println(err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("failed to open database")
	}
	dbQueries := database.New(db)

	// create a new state
	s := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	// create an instance of the commands struct and initialize the cmdmap
	cmds := commands{
		cmdmap: make(map[string]func(*state, command) error),
	}

	// register commands
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)

	// confirm the user input at least two args. Example: gator login
	if len(os.Args) < 2 {
		log.Fatal("you must input a command to use gator\n")
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
