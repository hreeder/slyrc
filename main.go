package main

import (
	"encoding/json"
	"fmt"
	"os"

	"log"

	"strings"

	"github.com/hreeder/slyrc/config"
	"github.com/hreeder/slyrc/irc"
	"github.com/nlopes/slack"
	thojirc "github.com/thoj/go-ircevent"
)

var ircConnections map[string]*irc.SlyrcIRCClient
var ready bool

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config := config.Configuration{}
	cfgErr := decoder.Decode(&config)
	if cfgErr != nil {
		fmt.Println(cfgErr)
		panic("Could not load config.json")
	}

	ircConnections = make(map[string]*irc.SlyrcIRCClient)
	ready = false

	slackAPI := slack.New(config.Slack.BotKey)
	slackLogger := log.New(os.Stdout, "slack: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(slackLogger)
	slackAPI.SetDebug(true)

	ircClient := thojirc.IRC(config.IRC.Nickname, config.IRC.Username)

	if config.IRC.Password != "" {
		ircClient.Password = config.IRC.Password
	}

	if config.IRC.TLS {
		ircClient.UseTLS = true
	}

	ircErr := ircClient.Connect(config.IRC.Server)
	if ircErr != nil {
		fmt.Println(ircErr)
		panic("Failed to connect to IRC")
	}
	fmt.Println("Connected to IRC")

	ircClient.AddCallback("001", func(e *thojirc.Event) {
		for _, pair := range config.Mappings {
			ircClient.Join(pair.IRC)
		}
		ready = true
	})

	ircClient.AddCallback("PRIVMSG", func(e *thojirc.Event) {
		if !strings.HasPrefix(e.Nick, "[") && !strings.HasSuffix(e.Nick, "]") {
			slackTargetChannel := GetChannel(config.Mappings, e.Arguments[0], "slack")
			msgParams := slack.NewPostMessageParameters()
			msgParams.Username = e.Nick
			msgParams.AsUser = false
			msgParams.IconURL = "http://i.imgur.com/2R9IRhz.png"
			slackAPI.PostMessage(slackTargetChannel, e.Message(), msgParams)
		}
	})

	fmt.Println("Added IRC Callbacks")

	fmt.Println("Starting Slack connection")
	HandleAndRunSlack(config, slackAPI, ircClient)
}
