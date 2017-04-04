package main

import (
	"encoding/json"
	"fmt"
	"os"

	"log"

	"github.com/hreeder/slyrc/config"
	"github.com/hreeder/slyrc/irc"
	"github.com/nlopes/slack"
	thojirc "github.com/thoj/go-ircevent"
)

var cfg config.Configuration
var ircConnections map[string]*irc.SlyrcIRCClient
var ready bool
var slackAPI *slack.Client
var ircClient *thojirc.Connection

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	cfg = config.Configuration{}
	cfgErr := decoder.Decode(&cfg)
	if cfgErr != nil {
		fmt.Println(cfgErr)
		panic("Could not load config.json")
	}

	ircConnections = make(map[string]*irc.SlyrcIRCClient)
	ready = false

	slackAPI = slack.New(cfg.Slack.BotKey)
	slackLogger := log.New(os.Stdout, "slack: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(slackLogger)
	slackAPI.SetDebug(false)

	ircClient = thojirc.IRC(cfg.IRC.Nickname, cfg.IRC.Username)

	if cfg.IRC.Password != "" {
		ircClient.Password = cfg.IRC.Password
	}

	if cfg.IRC.TLS {
		ircClient.UseTLS = true
	}

	ircErr := ircClient.Connect(cfg.IRC.Server)
	if ircErr != nil {
		fmt.Println(ircErr)
		panic("Failed to connect to IRC")
	}
	fmt.Println("Connected to IRC")

	ircClient.AddCallback("001", onConnect)

	ircClient.AddCallback("PRIVMSG", onPrivmsg)

	fmt.Println("Added IRC Callbacks")

	fmt.Println("Starting Slack connection")
	HandleAndRunSlack(slackAPI, ircClient)
}
