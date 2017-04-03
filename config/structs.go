package config

type ircConfiguration struct {
	Server   string `json:"server"`
	TLS      bool   `json:"tls"`
	Nickname string `json:"nickname"`
	Username string `json:"user"`
	Password string `json:"password"`
}

type slackConfiguration struct {
	APIKey string `json:"api_key"`
	BotKey string `json:"bot_key"`
}

// ChannelMapping represents a mapping between
// an IRC channel and a slack channel
type ChannelMapping struct {
	IRC   string `json:"irc"`
	Slack string `json:"slack"`
}

// Configuration models the data we expect to
// see come in from our config.json
type Configuration struct {
	IRC      ircConfiguration   `json:"irc"`
	Slack    slackConfiguration `json:"slack"`
	Mappings []ChannelMapping   `json:"mappings"`
	Timeout  int                `json:"timeout"`
}
