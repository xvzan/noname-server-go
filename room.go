package main

import (
	"fmt"
	"sync"
)

type Room struct {
	Key        string
	Owner      *Client
	Config     map[string]interface{}
	ServerMode bool
	UserCount  int
	// mu         sync.Mutex
}

var (
	roomList = make(map[string]*Room)
	roomMu   sync.Mutex
)

func CreateRoom(client *Client, key string) {

	if client.OnlineKey != key {
		return
	}
	defaultConfig := map[string]string{
		"mode":             "identity",
		"identity_mode":    "zhong",
		"double_character": "true",
		"number":           "8",
	}

	room := &Room{
		Key:        key,
		Owner:      client,
		Config:     map[string]interface{}{},
		ServerMode: client.ServerMode,
	}
	for k, v := range defaultConfig {
		room.Config[k] = v
	}

	roomMu.Lock()
	roomList[key] = room
	roomMu.Unlock()
	client.Room = room
	client.sendl(modifyMessage([]string{"createroom", key}))
}

func updateRooms() {
	rooms := getRoomList()
	clientslist := getClientList()
	clientsLock.Lock()
	for _, c := range clients {
		if c.Room == nil {
			c.sendl(modifyMessage(append([]interface{}{"updaterooms"}, rooms, clientslist)))
		}
	}
	clientsLock.Unlock()
}

func EnterRoom(client *Client, key string, nickname, avatar string) {
	defer fmt.Println("EnterRoom completed for", client.ID, "in room", key)
	client.Nickname = trimNickname(nickname)
	client.Avatar = avatar
	roomMu.Lock()

	room, ok := roomList[key]
	if !ok || room.Owner == nil {
		client.sendl(modifyMessage([]string{"enterroomfailed"}))
		return
	}
	client.Room = room
	roomMu.Unlock()

	if room.ServerMode && room.Owner._OnConfig == nil {
		room.Owner.sendl(modifyMessage([]string{"createroom", key}))
		room.Owner._OnConfig = client
		room.Owner.Nickname = client.Nickname
		room.Owner.Avatar = avatar
	} else if room.Config == nil || (room.Config["gameStarted"] == true &&
		(room.Config["observe"] == false || room.Config["observeReady"] == false)) {
		client.sendl(modifyMessage([]string{"enterroomfailed"}))
	} else {
		client.Owner = room.Owner
		room.Owner.sendl(modifyMessage([]string{"onconnection", client.ID}))
	}
	updateRooms()
}

func getRoomList() [][]interface{} {
	roomMu.Lock()
	defer roomMu.Unlock()

	for _, room := range roomList {
		room.UserCount = 0
	}

	clientsLock.Lock()
	defer clientsLock.Unlock()

	for _, client := range clients {
		if client.Room != nil && !client.ServerMode {
			client.Room.UserCount++
		}
	}

	result := make([][]interface{}, 0, len(roomList))
	for _, room := range roomList {
		if room.ServerMode {
			result = append(result, []interface{}{"server"})
		} else if room.Owner != nil && room.Config != nil {
			if room.UserCount == 0 && room.Owner != nil {
				room.Owner.sendl(modifyMessage([]string{"reloadroom"}))
			}
			entry := []interface{}{
				room.Owner.Nickname,
				room.Owner.Avatar,
				room.Config,
				room.UserCount,
				room.Key,
			}
			result = append(result, entry)
		}
	}
	return result
}
