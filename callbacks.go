package main

import (
	"strings"

	"fmt"

	"github.com/nlopes/slack"
	thojirc "github.com/thoj/go-ircevent"
)

func onPrivmsg(e *thojirc.Event) {
	if !strings.HasPrefix(e.Nick, "[") && !strings.HasSuffix(e.Nick, "]") {
		slackTargetChannel := GetChannel(cfg.Mappings, e.Arguments[0], "slack")
		msgParams := slack.NewPostMessageParameters()
		msgParams.Username = e.Nick
		msgParams.AsUser = false
		msgParams.IconURL = "http://i.imgur.com/2R9IRhz.png"
		slackAPI.PostMessage(slackTargetChannel, e.Message(), msgParams)
	}
}

func onConnect(e *thojirc.Event) {
	if cfg.IRC.NickservPassword != "" {
		ircClient.Privmsg("NICKSERV", fmt.Sprintf("IDENTIFY %s", cfg.IRC.NickservPassword))
	}

	for _, pair := range cfg.Mappings {
		ircClient.Join(pair.IRC)
	}
	ready = true
}
