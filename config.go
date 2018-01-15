package main

import (
	"encoding/json"
	"os"
)

// Configuration holds information about application configuration.
type Configuration struct {
	CloudletName         string
	CloudletIP           string
	CloudletDomain       string
	CloudletRegisterLink string
	ServiceRegisterLink  string

	Services []Service
}

type Service struct {
	Name   string
	Domain string
}

// Config holds configuration values.
var Config Configuration

// ReadConfigurationFile builds configuration instance from json config file.
// This method must be called at the beginning of main method.
func ReadConfigurationFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	var config Configuration
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	Config = config

	return nil
}
