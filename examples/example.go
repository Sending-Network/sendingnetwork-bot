package main

import (
	"bufio"
	"fmt"
	sdnclient "github.com/sending-network/sendingnetwork-bot"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
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

	outConfig, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}
	_ = os.WriteFile("config.yaml", outConfig, 0644)

	syncer := cli.Syncer.(*sdnclient.DefaultSyncer)
	syncer.OnEventType("m.room.message", func(ev *sdnclient.Event) {
		fmt.Println("Message: ", ev)
	})

	go func() {
		for {
			if err := cli.Sync(); err != nil {
				fmt.Println("Sync() returned ", err)
			}
			// Optional: Wait a period of time before trying to sync again.
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		processCommand(cli, scanner.Text())
	}
}

func processCommand(client *sdnclient.Client, command string) {
	parts := strings.Split(command, " ")
	if len(parts) < 2 || parts[0] != "room" {
		return
	}
	action := parts[1]
	switch action {
	case "list":
		resp, err := client.GetJoinedRooms()
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp.JoinedRooms)
		}
	case "create":
		roomName := parts[2]
		resp, err := client.CreateRoom(&sdnclient.ReqCreateRoom{Name: roomName})
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp.RoomID)
		}
	case "join":
		roomId := parts[2]
		resp, err := client.JoinRoom(roomId)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp.RoomID)
		}
	case "leave":
		roomId := parts[2]
		_, err := client.LeaveRoom(roomId)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println("leave success")
		}
	case "invite":
		roomId := parts[2]
		userId := parts[3]
		_, err := client.InviteUser(roomId, &sdnclient.ReqInviteUser{UserID: userId})
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println("invite success")
		}
	case "kick":
		roomId := parts[2]
		userId := parts[3]
		_, err := client.KickUser(roomId, &sdnclient.ReqKickUser{UserID: userId})
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println("kick success")
		}
	case "members":
		roomId := parts[2]
		resp, err := client.JoinedMembers(roomId)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp.Joined)
		}
	case "send":
		roomId := parts[2]
		msg := parts[3]
		resp, err := client.SendText(roomId, msg)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp.EventID)
		}
	case "state":
		roomId := parts[2]
		eventType := parts[3]
		stateKey := ""
		if len(parts) >= 5 {
			stateKey = parts[4]
		}
		resp, err := client.GetStateEvent(roomId, eventType, stateKey)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		} else {
			fmt.Println(resp)
		}
	}
}
