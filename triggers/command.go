package triggers

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/lavagetto/ircbot/acl"
	"github.com/lavagetto/ircbot/bot"
	hbot "github.com/whyrusleeping/hellabot"
	log "gopkg.in/inconshreveable/log15.v2"
)

/*
	Commands section
*/
type CommandClosure func(
	[]string,
	*hbot.Bot,
	*hbot.Message,
	*bot.Configuration,
	*sql.DB,
) bool

// Command encapsulates an irc command
type Command struct {
	// The command identifier - it will determine how
	// the command is called
	ID              string
	ArgumentsRegexp *regexp.Regexp
	HelpMsg         string
	privmsg         bool
	public          bool
	Action          CommandClosure
	Db              *sql.DB
	Configuration   *bot.Configuration
}

// Helper functions when initializing from ircbot.IrcBot.AddCommand
// Allows triggering the command via private message to the bot
func (cmd *Command) AllowPrivate() *Command {
	cmd.privmsg = true
	return cmd
}

// Allows triggering the command via public message in a channel
func (cmd *Command) AllowChannel() *Command {
	cmd.public = true
	return cmd
}

func (cmd *Command) SetHelp(msg string) *Command {
	cmd.HelpMsg = msg
	return cmd
}

func (cmd *Command) Arguments(argRegexp string) *Command {
	cmd.ArgumentsRegexp = regexp.MustCompile(argRegexp)
	return cmd
}

// Checks if we should act on the event.
func (cmd Command) isCommand(bot *hbot.Bot, m *hbot.Message) bool {
	if cmd.ID == "" {
		return false
	}
	// The action is triggered to private messages for !command
	// or public messages for !command
	// depending on if they're enabled or not
	// Please note that PRIVMSG in hellabot conventions is any message received
	// by the bot, either public or private.
	if m.Command != "PRIVMSG" {
		return false
	}
	maybeCommand := strings.Fields(m.Content)[0]
	if maybeCommand != "!"+cmd.ID {
		return false
	}
	// Do not accept commands in a channel if they're not public
	if !cmd.public && strings.HasPrefix(m.To, "#") {
		return false
	}
	// Do not accept private commands if they're not private
	if !cmd.privmsg && !strings.HasPrefix(m.To, "#") {
		return false
	}
	return true
}

// Checks if the sender/channel allow the action.
func (cmd Command) checkAcl(irc *hbot.Bot, m *hbot.Message) bool {
	acl, err := acl.GetACL(cmd.ID, cmd.Db, cmd.Configuration)
	if err != nil {
		// We log the issue, but we don't stop admins from being able to perform commands.
		log.Error("Couldn't fetch the ACLs", "error", err.Error())
	}
	if !acl.IsAllowed(m) {
		irc.Reply(m, "You're not allowed to perform this action.")
		return false
	} else {
		return true
	}

}

func (cmd Command) doAction(irc *hbot.Bot, m *hbot.Message) bool {
	var matches []string
	if cmd.ArgumentsRegexp != nil {
		args := strings.Join(strings.Fields(m.Content)[1:], " ")
		// Validate the content of the string
		matches = cmd.ArgumentsRegexp.FindStringSubmatch(args)
		if matches == nil {
			irc.Reply(m, "The command is not properly formatted.")
			irc.Reply(m, cmd.Help())
			return false
		}
	} else {
		matches = []string{cmd.ID}
	}
	return cmd.Action(matches[1:], irc, m, cmd.Configuration, cmd.Db)
}

func (cmd Command) Help() string {
	parameters := []string{fmt.Sprintf("!%s", cmd.ID)}
	for i, parameter := range cmd.ArgumentsRegexp.SubexpNames()[1:] {
		if parameter == "" {
			parameter = fmt.Sprintf("arg%d", i)
		}
		parameters = append(parameters, fmt.Sprintf("<%s>", parameter))
	}
	return fmt.Sprintf("%s. Format: %s", cmd.HelpMsg, strings.Join(parameters, " "))

}

func (cmd Command) Handle(irc *hbot.Bot, m *hbot.Message) bool {
	//log.Info("Handling message", "command", m.Command, "to", m.To, "content", m.Content)
	if cmd.isCommand(irc, m) && cmd.checkAcl(irc, m) {
		return cmd.doAction(irc, m)
	} else {
		return false
	}
}
