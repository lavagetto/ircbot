package triggers

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/lavagetto/ircbot/bot"

	log "gopkg.in/inconshreveable/log15.v2"

	hbot "github.com/whyrusleeping/hellabot"
)

// TriggerFunc is the format of the functions we expect.
// They depend on the irc bot, the message, the db and the configuration
type TriggerFunc func(*hbot.Bot, *hbot.Message, *sql.DB, *bot.Configuration) bool

type HelpHandler interface {
	Handle(*hbot.Bot, *hbot.Message) bool
	Help() string
}

// EvHandler is the basic structure for holding information about an
// irc trigger. For most interactive commands, where you want to define ACLs,
// input validation, etc. you should use Command instead.
type EvHandler struct {
	Handler TriggerFunc
	HelpMsg string
	Config  *bot.Configuration
	Db      *sql.DB
}

// Handle manages event hooks to see if they're appliable to the incoming request
func (ev EvHandler) Handle(irc *hbot.Bot, m *hbot.Message) bool {
	// Exactly like a common trigger
	return ev.Handler(irc, m, ev.Db, ev.Config)
}

func (ev EvHandler) Help() string {
	return ev.HelpMsg
}

// Registry is a container for all event handlers.
type Registry struct {
	// All the handlers, by ID
	handlers map[string]HelpHandler
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	var r Registry
	r.handlers = make(map[string]HelpHandler)
	return &r
}

// Register an handler. This is the basic interface you should use if you're not crating
// a proper command, but rather a trigger.
// For interactive commands, please use RegisterCommand below.
func (r *Registry) Register(id string, handler TriggerFunc, help string, db *sql.DB, c *bot.Configuration) error {
	if _, ok := r.handlers[id]; ok {
		msg := fmt.Sprintf("Cannot register handler with id '%s' twice", id)
		return errors.New(msg)
	}
	var h HelpHandler = EvHandler{
		Handler: handler,
		HelpMsg: help,
		Db:      db,
		Config:  c,
	}
	r.handlers[id] = h
	return nil
}

// RegisterCommand allows to register a full-featured IRC command.
func (r *Registry) RegisterCommand(command *Command) error {
	id := command.ID
	if _, ok := r.handlers[id]; ok {
		msg := fmt.Sprintf("Cannot register handler with id '%s' twice", id)
		return errors.New(msg)
	}
	var h HelpHandler = *command
	r.handlers[id] = h
	return nil
}

func (r *Registry) RegisterCommands(commands []*Command) error {
	for _, command := range commands {
		err := r.RegisterCommand(command)
		if err != nil {
			return err
		}
	}
	return nil
}

// Deregister removes one handler from the system.
func (r *Registry) Deregister(id string) {
	delete(r.handlers, id)
}

func (r *Registry) AddAll(b *bot.Bot, c *bot.Configuration) {
	for id, Handler := range r.handlers {
		log.Info("Registering handler", "id", id)
		b.Irc.AddTrigger(Handler)
	}
	r.addHelp(b, c)
}

// Help prints out the help for the registered commands
func (r *Registry) addHelp(b *bot.Bot, c *bot.Configuration) {
	b.Irc.AddTrigger(
		hbot.Trigger{
			Condition: func(bot *hbot.Bot, m *hbot.Message) bool {
				return m.Command == "PRIVMSG" && m.Content == "!help"
			},
			Action: func(bot *hbot.Bot, m *hbot.Message) bool {
				bot.Reply(m, fmt.Sprintf("%s - irc bot for handling outages", c.NickName))
				bot.Reply(m, "")
				bot.Reply(m, "Available commands:")
				bot.Reply(m, fmt.Sprintf("%-16s%s\n", "!help", "Prints this message"))
				var handlers_help = make([]string, 0, len(r.handlers))
				// get the help messages for all handlers that have one.
				for name, handler := range r.handlers {
					help_msg := handler.Help()
					// Some commands might not have an help message by design...
					if help_msg != "" {
						handlers_help = append(handlers_help, fmt.Sprintf("%-16s%s\n", name, help_msg))
					}
				}
				// We want a sorted output
				sort.Strings(handlers_help)
				for _, msg := range handlers_help {
					bot.Reply(m, msg)
				}
				return true
			},
		},
	)
}
