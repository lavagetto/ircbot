package bot

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Configuration holds all the configuration of
// the bot
type Configuration struct {
	// Name of the server you're connecting to
	ServerName string `json:"server"`
	// Server TCP port
	ServerPort uint `json:"port"`
	// set to true if you want to connect via TLS
	UseTLS bool `json:"use_tls"`
	// Set to true to use SASL auth
	UseSASL bool `json:"use_sasl"`
	// Nickname
	NickName string `json:"nick"`
	// NickServ password for the given nickname
	Password string `json:"password"`
	// Array of chat channels to join
	Channels []string `json:"channels"`
	// Which of these channels are considered public (so a reduced amount of data)
	// will be reported.
	PublicChannels []string `json:"public_channels"`
	// DSN of the database connection.
	DbDsn string `json:"db_dsn"`
	// The nicknames of the admins of the bot.
	// You should take care of ensure their nicknames are
	// protected
	Admins []string `json:"admins"`
	// Auth credentials file for access to GDocs
}

// GetConfig initializes a configuration object
// from reading a properly formatted json file
func GetConfig(fileName string) (*Configuration, error) {
	config := Configuration{
		ServerName: "irc.libera.chat",
		ServerPort: 6697,
		UseTLS:     true,
		UseSASL:    true,
		NickName:   "IrcBot",
		Channels:   []string{"#somechannel"},
		DbDsn:      "sqlite3://file:ircbot.db?cache=shared",
	}
	if fileName == "" {
		return &config, nil
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, err
}

// GetServerString gives you a host:port string of the server to connect to.
func (c *Configuration) GetServerString() string {
	return fmt.Sprintf("%s:%d", c.ServerName, c.ServerPort)
}

// IsPublicChannel tells you if a channel is public or not.
func (c *Configuration) IsPublicChannel(channel string) bool {
	sort.Strings(c.PublicChannels)
	i := sort.SearchStrings(c.PublicChannels, channel)
	return i < len(c.PublicChannels)
}
