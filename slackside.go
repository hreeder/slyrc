package main

import (
	"fmt"

	"github.com/hreeder/slyrc/irc"
	"github.com/nlopes/slack"
	thojirc "github.com/thoj/go-ircevent"
)

// HandleAndRunSlack will set up the RTM instance,
// Beginning the websocket and start handling incoming
// events from the channel
func HandleAndRunSlack(slackAPI *slack.Client, ircClient *thojirc.Connection) {
	rtm := slackAPI.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		// fmt.Printf("Event: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Nothing, ignore
		case *slack.ConnectedEvent:
			fmt.Println("Info: ", ev.Info)
			fmt.Println("Connection Counter: ", ev.ConnectionCount)

		case *slack.PresenceChangeEvent:
			fmt.Println("Type: ", ev.Type)
			user, userErr := slackAPI.GetUserInfo(ev.User)
			if userErr != nil {
				fmt.Println("Could not get user info for", ev.User)
				continue
			}

			if user.IsBot {
				fmt.Println(user.Name, "is a bot, ignoring.")
				continue
			}
			fmt.Println("User:", ev.User, "-", user.Name)
			fmt.Println("Presence:", ev.Presence)
			if ev.Presence == "active" {
				if _, ok := ircConnections[user.Name]; !ok {
					// this is where we connect to IRC
					ircConnections[user.Name] = irc.MakeIRCClient(user.Name, cfg)
					ircConnections[user.Name].Connect()
				}
			}

		case *slack.MessageEvent:
			if ready == false {
				continue
			}

			if ev.SubType == "bot_message" ||
				ev.SubType == "message_changed" {
				continue
			}

			// fmt.Printf("Message: %v\n", ev)
			targetIRCChannel := GetChannel(cfg.Mappings, ev.Channel, "irc")
			user, _ := slackAPI.GetUserInfo(ev.User)
			if targetIRCChannel != "" {
				if _, ok := ircConnections[user.Name]; !ok {
					// this is where we connect to IRC
					ircConnections[user.Name] = irc.MakeIRCClient(user.Name, cfg)
					ircConnections[user.Name].Connect()
					ircConnections[user.Name].JoinChannel(targetIRCChannel)
					ircConnections[user.Name].SendMessage(targetIRCChannel, ev.Text)
					continue
				} else if _, ok := ircConnections[user.Name]; ok {
					ircConnections[user.Name].JoinChannel(targetIRCChannel)
					ircConnections[user.Name].SendMessage(targetIRCChannel, ev.Text)
					continue
				}
			}

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)
			// expire any old clients
			for key, client := range ircConnections {
				toDelete := client.TryExpire()
				if toDelete {
					delete(ircConnections, key)
				}
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

			// default:
			// 	fmt.Printf("Unknown Type: %v", reflect.TypeOf(msg.Data))
			// fmt.Printf("Unknown: %v\n", ev)
		}
	}
}
