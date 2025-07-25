package main

import (
	"math/rand"
	"strconv"
	"strings"
)

var (
	bannedKeys     = []string{"badkey1", "malicious"}
	bannedKeyWords = []string{"spam", "illegal"}
)

func generateID() string {
	return strconv.Itoa(1000000000 + rand.Intn(9000000000))
}

func trimNickname(nick string) string {
	if len(nick) > 30 {
		return nick[:30]
	}
	if nick == "" {
		return "无名玩家"
	}
	return nick
}

func isBanned(content string) bool {
	for _, word := range bannedKeyWords {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}

func encodeMessage(parts ...string) string {
	return strings.Join(parts, ":")
}

func isKeyBanned(key string) bool {
	for _, banned := range bannedKeys {
		if key == banned {
			return true
		}
	}
	return false
}

func buildMessage(label string, slices ...[]string) []string {
	result := []string{label}
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func serializeConfig(config map[string]string) []interface{} {
	result := make([]interface{}, 0)
	for k, v := range config {
		result = append(result, k+"="+v)
	}
	return result
}

func modifyMessage[T any](items []T) []interface{} {
	res := make([]interface{}, len(items))
	for i, v := range items {
		res[i] = v
	}
	return res
}
