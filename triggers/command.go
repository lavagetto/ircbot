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
	map[string]string,
	*hbot.Bot,
	*hbot.Message,
	*bot.Configuration,
	*sql.DB,
) bool

type argsCallback func(
	*hbot.Message,
) string

type CommandArgument struct {
	validator       *regexp.Regexp
	defaultCallback argsCallback
}

func (c *CommandArgument) SetValidator(reg string) {
	c.validator = regexp.MustCompile(reg)
}

// Set default value
func (c *CommandArgument) Default(value string) {
	c.defaultCallback = func(*hbot.Message) string {
		return value
	}
}

func (c *CommandArgument) Validate(value string) error {
	if c.validator.FindStringIndex(value) != nil {
		return nil
	}

	return fmt.Errorf("the value %s doesn't match the regexp %s", value, c.validator.String())
}

func (c *CommandArgument) Get(value string, m *hbot.Message) string {
	if value != "" {
		return value
	}
	if c.defaultCallback != nil {
		return c.defaultCallback(m)
	}
	return ""
}

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
	parameters      map[string]*CommandArgument
	paramOder       []string
}

func (cmd *Command) InitParams() {
	if cmd.parameters == nil {
		cmd.parameters = make(map[string]*CommandArgument)
	}
	if cmd.paramOder == nil {
		cmd.paramOder = make([]string, 0)
	}
}

func (cmd *Command) addParameter(name string, c *CommandArgument) {
	cmd.parameters[name] = c
	cmd.paramOder = append(cmd.paramOder, name)
}

func (cmd *Command) AddParameter(name string, regex string) *Command {
	c := &CommandArgument{}
	c.SetValidator(regex)
	cmd.addParameter(name, c)
	return cmd
}

func (cmd *Command) AddParameterWithDefault(name string, regex string, defaultValue string) *Command {
	c := &CommandArgument{}
	c.SetValidator(fmt.Sprintf("(%s)?", regex))
	c.Default(defaultValue)
	cmd.addParameter(name, c)
	return cmd
}

func (cmd *Command) AddParameterWithDefaultCb(name string, regex string, defaultCb argsCallback) *Command {
	c := &CommandArgument{}
	c.SetValidator(fmt.Sprintf("(%s)?", regex))
	c.defaultCallback = defaultCb
	cmd.addParameter(name, c)
	return cmd
}

func (cmd *Command) Parameter(name string) *CommandArgument {
	if _, ok := cmd.parameters[name]; !ok {
		cmd.addParameter(name, &CommandArgument{})
	}
	return cmd.parameters[name]
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

// Add all arguments in a single regexp
// the regexp must use named grouped parameters.
// Deprecated: you should use the AddParameter* functions
// instead, which provide more advanced handling.
func (cmd *Command) Arguments(argRegexp string) *Command {
	if len(cmd.parameters) > 0 {
		panic("A command can't have both named parameters and Arguments() calls")
	}
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
	args, err := cmd.parseMessage(m)
	if err != nil {
		irc.Reply(m, err.Error())
	}
	return cmd.Action(args, irc, m, cmd.Configuration, cmd.Db)
}

func (cmd Command) parseMessage(m *hbot.Message) (map[string]string, error) {
	args := make(map[string]string)
	rawArgs := strings.Fields(m.Content)[1:]
	numRawArgs := len(rawArgs)
	if cmd.ArgumentsRegexp != nil {
		arg_names := cmd.ArgumentsRegexp.SubexpNames()
		argsStr := strings.Join(rawArgs, " ")
		// Validate the content of the string
		matches := cmd.ArgumentsRegexp.FindStringSubmatch(argsStr)
		if matches == nil {
			return args, fmt.Errorf("the command is not properly formatted")
		}
		for i, match := range matches[1:] {
			name := arg_names[i]
			args[name] = match
		}
		return args, nil
	}
	for idx, param := range cmd.paramOder {
		c := cmd.Parameter(param)
		var value string
		// A value was provided
		if numRawArgs > idx {
			err := c.Validate(rawArgs[idx])
			if err != nil {
				return args, err
			}
			value = c.Get(rawArgs[idx], m)
		} else {
			value = c.Get("", m)
		}
		// If no value was provided, and no default was provided, return an error
		if value == "" {
			return args, fmt.Errorf("no value provided for parameter %s and no default available", param)
		}
		args[param] = value
	}
	return args, nil
}

func (cmd Command) Help() string {
	// don't show help if none was provided.
	if cmd.HelpMsg == "" {
		return ""
	}
	parameters := []string{fmt.Sprintf("!%s", cmd.ID)}
	if cmd.ArgumentsRegexp != nil {
		for i, parameter := range cmd.ArgumentsRegexp.SubexpNames()[1:] {
			if parameter == "" {
				parameter = fmt.Sprintf("arg%d", i)
			}
			parameters = append(parameters, fmt.Sprintf("<%s>", parameter))
		}
	}
	for _, parameter := range cmd.paramOder {
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
