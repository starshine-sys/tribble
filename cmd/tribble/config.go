// SPDX-License-Identifier: AGPL-3.0-only
package main

import (
	"log"
	"os"

	"github.com/starshine-sys/tribble/common"
	"gopkg.in/yaml.v2"
)

func getConfig() (config *common.BotConfig) {
	config = &common.BotConfig{}

	configFile, err := os.ReadFile("data/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Loaded configuration file.")

	// override some values with env variables
	if s := os.Getenv("DATABASE_URL"); s != "" {
		config.DatabaseURL = s
	}

	return config
}
