package main

import (
	"os"

	sdnclient "github.com/sending-network/sendingnetwork-bot"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func main() {
	configData, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	config := sdnclient.Config{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("error: %v", err)
	}

	cli, err := sdnclient.NewClient(&config)
	if err != nil {
		log.Fatal(err)
	}

	roomID := "!HMixe2dD3IcLXN9g-@sdn_e199304b4349f43978575ed4c48d0664513e63fb:e199304b4349f43978575ed4c48d0664513e63fb"
	respJoinRoom, err := cli.JoinRoom(roomID)
	if err != nil {
		log.Errorf("JoinRoom fail: %v", err)
		return
	}
	log.Infof("JoinRoom: %v", respJoinRoom)
}
