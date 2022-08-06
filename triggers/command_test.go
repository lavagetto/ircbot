package triggers

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/lavagetto/ircbot/bot"

	hbot "github.com/whyrusleeping/hellabot"
)

type newCommandParams struct {
	name        string
	regexString string
	help        string
	public      bool
	private     bool
	action      CommandClosure
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

func TestCommandSimple(t *testing.T) {
	expected := map[string]string{"param": "what is there"}
	commandVars := newCommandParams{
		"test_command",
		`(?P<param>\w+)\s+(.*)$`,
		"Test command",
		true,
		true,
		getVerifyArgs(expected, t),
	}
	// TODO: implement this
	fmt.Println(commandVars)
}
