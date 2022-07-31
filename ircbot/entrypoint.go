package ircbot

import (
	"database/sql"

	"github.com/lavagetto/ircbot/bot"
	"github.com/lavagetto/ircbot/triggers"
	"github.com/lavagetto/ircbot/utils"
	hbot "github.com/whyrusleeping/hellabot"
	log "gopkg.in/inconshreveable/log15.v2"

	_ "github.com/mattn/go-sqlite3"
)

// This is the entrypoint for ircbot.
type IrcBot struct {
	// Holds the configuration file name
	conf *bot.Configuration
	// The bot structure
	bot *bot.Bot
	// the Registry we can add commands to
	registry *triggers.Registry
	// List of irc commands added via the AddCommand interface
	ircCommands []*triggers.Command
}

// Initializes the bot.
// Pass it a configfile and you'll have in return
// a functioning instance of an IRC bot.
func Init(configFile string) (*IrcBot, error) {
	conf, err := bot.GetConfig(configFile)
	if err != nil {
		log.Info("Could not open configuration file")
	}
	bbot, err := bot.NewBot(conf)
	if err != nil {
		return nil, err
	}
	// Create a new command registry
	registry := triggers.NewRegistry()
	// A simple event handler with no command associated
	// stores the topic of a channel when it changes.
	registry.Register("store_topic", utils.StoreTopic, "", bbot.DB, conf)

	irc := &IrcBot{
		conf:        conf,
		bot:         bbot,
		registry:    registry,
		ircCommands: make([]*triggers.Command, 0),
	}
	irc.AddBuiltins()
	return irc, nil
}

func (irc *IrcBot) RegisterCommands(commands []*triggers.Command) error {
	return irc.registry.RegisterCommands(commands)
}

func (irc *IrcBot) Run() {
	// Register all commands that have been added,
	// We need to only register commands here because
	// the registry dereferences them.
	for _, command := range irc.ircCommands {
		irc.registry.RegisterCommand(command)
	}
	irc.registry.AddAll(irc.bot, irc.Config())
	defer irc.DB().Close()
	irc.bot.Irc.Run()
}

// Returns the database handle. Useful in commands.
func (irc *IrcBot) DB() *sql.DB {
	return irc.bot.DB
}

// Returns the configuration
func (irc *IrcBot) Config() *bot.Configuration {
	return irc.conf
}

// Reply to a message via IRC
func (irc *IrcBot) Reply(m *hbot.Message, what string) {
	irc.bot.Irc.Reply(m, what)
}

// get the hellabot logger
func (irc *IrcBot) Logger() log.Logger {
	return irc.bot.Irc.Logger
}

// Adds a non-configured command to the registry, that can be then configured.
func (irc *IrcBot) AddCommand(name string, action CommandAction) *triggers.Command {
	CommandClosure := func(args []string, bot *hbot.Bot, m *hbot.Message, c *bot.Configuration, db *sql.DB) bool {
		return action(args, m, irc)
	}
	c := &triggers.Command{
		ID:            name,
		Action:        CommandClosure,
		Db:            irc.DB(),
		Configuration: irc.Config(),
	}
	irc.ircCommands = append(irc.ircCommands, c)
	return c
}

func (irc *IrcBot) AddBuiltins() {
	irc.AddCommand("sing", sing).AllowChannel().AllowPrivate().Arguments("").SetHelp("Sings a nice tune.")
	irc.addAclCommand("acl_add", "Adds the ability for a command to be used by a single user or in a channel", addACL)
	irc.addAclCommand("acl_remove", "Removes a user/channel from the ACL", removeAcl)
	irc.addAclCommand("acl_get", "Gets the defined ACLs for a command", readAcl)
	irc.AddCommand("passwd", changePass).Arguments("(?P<new_password>\\S+)\\s*$").AllowPrivate().SetHelp("Changes the nickserv password.")
}

func (irc *IrcBot) addAclCommand(name string, help string, callback CommandAction) {
	arg := "(?P<command>\\S+)\\s+(?P<nick_or_chan>\\S+)\\s*$"
	if name == "acl_get" {
		arg = "(?P<command>\\S+)\\s*$"
	}
	irc.AddCommand(name, callback).AllowPrivate().Arguments(arg).SetHelp(help)
}

type CommandAction func(
	[]string,
	*hbot.Message,
	*IrcBot,
) bool
