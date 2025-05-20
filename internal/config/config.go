package config

import (
	"encoding/json"
	"os"
)

// Export a Config struct that represents the JSON file structure including struct tags

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

// Export a Read function that reads the JSON file at ~/.gatorconfig.json and returns a Config struct -

func Read() (Config, error) {
	println("Getting config file path...")
	fullpath, err := GetConfigFilePath()
	println("path: ", fullpath)
	if err != nil {
		return Config{}, err
	}
	file, err := os.Open(fullpath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := Config{}
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// it should read the file from the HOME directory, the decode the JSON string into a new Config struct -
// use os.UserHomeDir()

// Export a SetUser method on the Config struct that writes the config struct to the JSON file after setting the current_user_name field
func (cfg *Config) SetUser(username string) error {
	// set the CurrentUserName field of the Config struct
	cfg.CurrentUserName = username
	// write the Config struct to the JSON file
	write(*cfg)
	return nil
}

func write(cfg Config) error {
	// retrieve the full filepath
	fullpath, err := GetConfigFilePath()
	if err != nil {
		return err
	}
	// create the json file
	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()
	// encode the json data
	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}

func GetConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	fullpath := home + "/.gatorconfig.json"
	return fullpath, nil
}
