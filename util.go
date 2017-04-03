package main

import "github.com/hreeder/slyrc/config"

// GetChannel will return the paired channel for the given input
func GetChannel(mappings []config.ChannelMapping, known string, direction string) string {
	for _, mapping := range mappings {
		if direction == "slack" && known == mapping.IRC {
			return mapping.Slack
		} else if direction == "irc" && known == mapping.Slack {
			return mapping.IRC
		}
	}
	return ""
}
