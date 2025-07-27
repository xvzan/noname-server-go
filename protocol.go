package main

import (
	"bytes"
	"encoding/json"
)

func handleMessage(c *Client, msg []byte) {
	if bytes.Equal(msg, []byte("heartbeat")) {
		c.Mutex.Lock()
		c.Beat = false
		c.Mutex.Unlock()
		// log.Println("[心跳]", c.ID, "已收到心跳包")
		return
	}

	// if len(message) == 0 {
	// 	c.sendl(modifyMessage([]string{"denied", "banned"}))
	// 	return
	// }

	if c.Owner != nil {
		c.Owner.sendl(modifyMessage(append([]interface{}{"onmessage"}, c.ID, string(msg))))
		return
	}

	var message []interface{}
	unmarshalerr := json.Unmarshal(msg, &message)
	if unmarshalerr != nil {
		// log.Println(unmarshalerr)
		return
	}
	// log.Println("[接收到消息]", message)

	if message[0] == "server" && len(message) >= 2 {
		cmd := message[1]
		switch cmd {
		case "create":
			if len(message) >= 5 {
				key := message[2].(string)
				nickname := message[3].(string)
				avatar := message[4].(string)
				c.OnlineKey = key
				CreateRoom(c, key)
				c.Nickname = trimNickname(nickname)
				c.Avatar = avatar
			}
		case "enter":
			if len(message) >= 5 {
				key := message[2].(string)
				nickname := message[3].(string)
				avatar := message[4].(string)
				EnterRoom(c, key, nickname, avatar)
			}
		case "send":
			if len(message) >= 4 {
				targetID := message[2].(string)
				message := message[3].(string)
				c.sendTo(targetID, message)
			}
		case "close":
			if len(message) >= 3 {
				targetID := message[2].(string)
				c.closeClient(targetID)
			}
			updateClients()
		case "config":
			if len(message) >= 3 {
				room := c.Room
				if room != nil && room.Owner == c {
					if room.ServerMode {
						room.ServerMode = false
					}
					if c._OnConfig != nil {
						c._OnConfig = nil
						clientsLock.Lock()
						if clients[c._OnConfig.ID] == nil {
							c._OnConfig.Owner = c
							c.sendl(modifyMessage([]string{"onconnection", c._OnConfig.ID}))
						}
						clientsLock.Unlock()
					}
					if config, ok := message[2].(map[string]interface{}); ok {
						room.Config = config
					}
				}
			}
			updateRooms()
		case "key":
			if len(message) >= 3 {
				if strSlice, ok := message[2].([]string); ok {
					if isKeyBanned(strSlice[0]) {
						bannedIps[c.Conn.RemoteAddr().String()] = true
						// log.Println("IP banned:", c.Conn.RemoteAddr().String())
						c.Conn.Close()
						return
					}
					c.OnlineKey = strSlice[0]
				}
			}
		case "events":
			if len(message) >= 5 {
				cfg := map[string]string{
					"utc":      message[2].(string),
					"day":      message[3].(string),
					"hour":     "xx",
					"content":  message[4].(string),
					"nickname": c.Nickname,
					"avatar":   c.Avatar,
				}
				CreateOrUpdateEvent(c, cfg, c.OnlineKey, "create")
			} else if len(message) >= 4 {
				eventID := message[2].(string)
				mode := message[3].(string)
				JoinOrLeaveEvent(c, eventID, c.OnlineKey, mode)
			}
		case "status":
			if len(message) >= 3 {
				status := message[2]
				c.Status = status.(string)
			}
			updateClients()
		case "changeAvatar":
			if len(message) >= 4 {
				c.Nickname = trimNickname(message[2].(string))
				c.Avatar = message[3].(string)
			}
			updateClients()
		default:
			c.sendl(modifyMessage([]string{"denied", "unknowncommand"}))
		}
	}
}
