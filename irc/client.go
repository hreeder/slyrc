package irc

import (
	"fmt"
	"time"

	config "github.com/hreeder/slyrc/config"
	thojirc "github.com/thoj/go-ircevent"
)

const (
	stateDisconnected = iota
	stateConnecting   = iota
	stateConnected    = iota
)

type pendingMessage struct {
	channel string
	message string
}

// SlyrcIRCClient represents an IRC client for Slyrc
type SlyrcIRCClient struct {
	state       int
	innerClient *thojirc.Connection
	config      config.Configuration

	lastEvent int64

	pendingChannelJoins    []string
	pendingChannelMessages []*pendingMessage
}

// MakeIRCClient will return a new IRC client ready to run
func MakeIRCClient(username string, config config.Configuration) *SlyrcIRCClient {
	client := &SlyrcIRCClient{}

	nick := fmt.Sprintf("[%s]", username)
	ircUsername := fmt.Sprintf("%s via Slyrc Slack Bridge", username)

	client.innerClient = thojirc.IRC(nick, ircUsername)
	client.state = stateDisconnected
	client.config = config

	if config.IRC.Password != "" {
		client.innerClient.Password = config.IRC.Password
	}

	if config.IRC.TLS {
		client.innerClient.UseTLS = true
	}

	return client
}

// Connect will start the connection happening
func (cl *SlyrcIRCClient) Connect() {
	cl.state = stateConnecting
	ircErr := cl.innerClient.Connect(cl.config.IRC.Server)
	if ircErr != nil {
		fmt.Println(ircErr)
	}

	cl.innerClient.AddCallback("001", cl.onConnect)
	// handle 331 - RPL_NOTOPIC and 332 - RPL_TOPIC
	cl.innerClient.AddCallback("331", cl.onJoinChannel)
	cl.innerClient.AddCallback("332", cl.onJoinChannel)

	cl.updateTimestamp()
}

// TryExpire will attempt to shut down this IRC client
func (cl *SlyrcIRCClient) TryExpire() bool {
	now := time.Now().UTC().Unix()
	if now-int64(cl.config.Timeout) > cl.lastEvent {
		cl.innerClient.Disconnect()
		return true
	}
	return false
}

func (cl *SlyrcIRCClient) updateTimestamp() {
	cl.lastEvent = time.Now().UTC().Unix()
}

func (cl *SlyrcIRCClient) onConnect(ev *thojirc.Event) {
	cl.state = stateConnected
	for _, pending := range cl.pendingChannelJoins {
		cl.JoinChannel(pending)
	}
	cl.pendingChannelJoins = nil
}

func (cl *SlyrcIRCClient) onJoinChannel(ev *thojirc.Event) {
	for _, pending := range cl.pendingChannelMessages {
		fmt.Println("Pending:", pending.channel, "Channel:", ev.Arguments[1])
		if pending.channel == ev.Arguments[1] {
			cl.SendMessage(pending.channel, pending.message)
		}
	}
	cl.pendingChannelMessages = nil
}

// JoinChannel will join a corresponding IRC channel
func (cl *SlyrcIRCClient) JoinChannel(channel string) {
	if cl.state == stateConnected {
		cl.innerClient.Join(channel)
	} else {
		cl.pendingChannelJoins = append(cl.pendingChannelJoins, channel)
	}

	cl.updateTimestamp()
}

// SendMessage will send a message when ready
func (cl *SlyrcIRCClient) SendMessage(destination string, message string) {
	if cl.state == stateConnected {
		cl.innerClient.Privmsg(destination, message)
	} else {
		p := &pendingMessage{channel: destination, message: message}
		cl.pendingChannelMessages = append(cl.pendingChannelMessages, p)
	}

	cl.updateTimestamp()
}
