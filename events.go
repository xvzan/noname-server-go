package main

import (
	"strconv"
	"sync"
	"time"
)

type Event struct {
	ID       string
	UTC      int64
	Day      string
	Hour     string
	Content  string
	Creator  string
	Nickname string
	Avatar   string
	Members  []string
}

var (
	events     []Event
	eventsLock sync.Mutex
	maxEvents  = 20
)

func checkExpiredEvents() {
	now := time.Now().UnixMilli()
	eventsLock.Lock()
	defer eventsLock.Unlock()

	filtered := []Event{}
	for _, e := range events {
		if e.UTC > now {
			filtered = append(filtered, e)
		}
	}
	events = filtered
}

func CreateOrUpdateEvent(client *Client, cfg map[string]string, id string, mode string) {
	eventsLock.Lock()
	defer eventsLock.Unlock()

	if isBanned(cfg["content"]) {
		client.sendl(modifyMessage([]string{"eventsdenied", "ban"}))
		return
	}

	now := time.Now().UnixMilli()
	eventUTC, _ := strconv.ParseInt(cfg["utc"], 10, 64)
	if len(events) >= maxEvents {
		client.sendl(modifyMessage([]string{"eventsdenied", "total"}))
		return
	} else if eventUTC <= now {
		client.sendl(modifyMessage([]string{"eventsdenied", "time"}))
		return
	}

	newEvent := Event{
		ID:       generateID(),
		UTC:      eventUTC,
		Day:      cfg["day"],
		Hour:     cfg["hour"],
		Content:  cfg["content"],
		Creator:  id,
		Nickname: trimNickname(cfg["nickname"]),
		Avatar:   cfg["avatar"],
		Members:  []string{id},
	}
	events = append([]Event{newEvent}, events...)
	updateEventsBroadcast()
}

func JoinOrLeaveEvent(client *Client, eventID string, userID string, mode string) {
	eventsLock.Lock()
	defer eventsLock.Unlock()

	var found bool
	for i, e := range events {
		if e.ID == eventID {
			found = true
			if mode == "join" {
				if !contains(e.Members, userID) {
					e.Members = append(e.Members, userID)
					client.sendl(modifyMessage([]string{"eventjoined", eventID}))
				} else {
					client.sendl(modifyMessage([]string{"eventjoined", eventID, "already"}))
				}
			} else if mode == "leave" {
				if contains(e.Members, userID) {
					e.Members = remove(e.Members, userID)
					client.sendl(modifyMessage([]string{"eventleft", eventID}))
					if len(e.Members) == 0 {
						events = append(events[:i], events[i+1:]...)
					}
				} else {
					client.sendl(modifyMessage([]string{"eventleft", eventID, "notfound"}))
				}
			}
			break
		}
	}
	if !found {
		client.sendl(modifyMessage([]string{"eventnotfound", eventID}))
	}
	updateEventsBroadcast()
}

func updateEventsBroadcast() {
	checkExpiredEvents()
	clientsLock.Lock()
	defer clientsLock.Unlock()
	for _, c := range clients {
		if c.Room == nil {
			c.sendl(modifyMessage(append([]interface{}{"updateevents"}, serializeEvents())))
		}
	}
}

func serializeEvents() []interface{} {
	result := make([]interface{}, 0, len(events))
	for _, e := range events {
		eventObj := map[string]interface{}{
			"id":       e.ID,
			"utc":      e.UTC,
			"day":      e.Day,
			"hour":     e.Hour,
			"content":  e.Content,
			"creator":  e.Creator,
			"nickname": e.Nickname,
			"avatar":   e.Avatar,
			"members":  e.Members,
		}
		result = append(result, eventObj)
	}
	return result
}

func getEventList() []interface{} {
	checkExpiredEvents()
	return serializeEvents()
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func remove(slice []string, s string) []string {
	result := []string{}
	for _, v := range slice {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}
