package fxConsole

import (
	"encoding/json"
	"os"
	"fmt"
	"io/ioutil"
)

type ConsoleConfig struct {
	Username string			`json:"username"`
	Authorization string 	`json:"authorization"`
	DaemonID string 		`json:"daemon_id"`
}

func (conf *ConsoleConfig) SaveConfig() {
	json_data, _ := json.Marshal(conf)
	err := ioutil.WriteFile(FlaxtonConfigFile, json_data, 0666)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}

func (conf *ConsoleConfig) LoadConfig() {
	file, _ := os.Open(FlaxtonConfigFile)
	decoder := json.NewDecoder(file)
	err := decoder.Decode(conf)
	if err != nil {
		fmt.Println("Error parsing coniguration file ", FlaxtonConfigFile)
		os.Exit(1)
	}
}