package triggers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/lavagetto/ircbot/bot"

	_ "github.com/mattn/go-sqlite3"
	hbot "github.com/whyrusleeping/hellabot"
	sx "gopkg.in/sorcix/irc.v2"
)

func forgeMsg(content string) *hbot.Message {
	prf := sx.Prefix{Name: "me"}
	m := sx.Message{Command: "PRIVMSG", Prefix: &prf}
	return &hbot.Message{Message: &m, Content: content, To: "ircbot"}
}

func getsql() *sql.DB {
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	return db
}
func getConfig() *bot.Configuration {
	return &bot.Configuration{
		ServerName: "irc.libera.chat",
		ServerPort: 6697,
		UseTLS:     true,
		UseSASL:    true,
		NickName:   "IrcBot",
		Channels:   []string{"#somechannel"},
		DbDsn:      "sqlite3://file:ircbot.db?cache=shared",
		Admins:     []string{"me"},
	}
}

func getVerifyArgs(expected map[string]string, t *testing.T) CommandClosure {
	return func(args map[string]string, irc *hbot.Bot, m *hbot.Message, c *bot.Configuration, db *sql.DB) bool {
		for name, arg := range args {
			if expected[name] != arg {
				t.Errorf("%s != %s at args[%s]", expected[name], arg, name)
			}
		}
		return true
	}
}

func testCommand(exp map[string]string, t *testing.T) *Command {
	c := Command{
		ID:            "test_command",
		Action:        getVerifyArgs(exp, t),
		Db:            getsql(),
		Configuration: getConfig(),
	}
	c.InitParams()
	return &c
}

func TestCommandArgs(t *testing.T) {
	irc := &hbot.Bot{}
	expected := map[string]string{"param": "what"}
	c := testCommand(expected, t)
	c.AddParameter("param", `\w+`).AllowPrivate()
	m := forgeMsg("!test_command what")
	c.Handle(irc, m)
}

func TestCommandNotAuthorized(t *testing.T) {
	c := testCommand(nil, t)
	m := forgeMsg("!test_command what")
	c.AddParameter("param", `\w+`).AllowPrivate()
	m.Name = "another"
	// This will fail if the callback is ever called
	// as we're comparing args to a nil map
	c.Handle(&hbot.Bot{}, m)
}

func TestCommandDefault(t *testing.T) {
	expected := map[string]string{"param": "what"}
	c := testCommand(expected, t)
	m := forgeMsg("!test_command")
	c.AddParameterWithDefault("param", `\w+`, "what").AllowPrivate()
	c.Handle(&hbot.Bot{}, m)
}

func TestCommandDefaultCb(t *testing.T) {
	expected := map[string]string{"param": "to:ircbot"}
	cb := func(m *hbot.Message) string {
		return fmt.Sprintf("to:%s", m.To)
	}
	c := testCommand(expected, t)
	m := forgeMsg("!test_command")
	c.AddParameterWithDefaultCb("param", `\w+`, cb).AllowPrivate()
	c.Handle(&hbot.Bot{}, m)
}
