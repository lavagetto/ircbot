package ircbot

import (
	"fmt"
	"time"

	"github.com/lavagetto/ircbot/acl"
	hbot "github.com/whyrusleeping/hellabot"
)

/*
 Builtin commands
**/

/*

No bot is a real bot without a rickroll

 **/

// I know there is a copyright on these. I consider this fair use.
// If you're not Rick Astley, don't bother opening a github issue
// to remove them, Thanks.
// If you're Rick Astley, please open the issue. Thanks, you've just
// made me twitter-famous.
var lyrics = []string{
	"Never gonna give you up",
	"Never gonna let you down",
	"Never gonna run around and desert you",
	"Never gonna make you cry",
	"Never gonna say goodbye",
	"Never gonna tell a lie and hurt you",
}

func sing(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	for _, line := range lyrics {
		irc.Reply(m, line)
		time.Sleep(800 * time.Millisecond)
	}
	return true
}

func porcessAclParams(args map[string]string, m *hbot.Message, irc *IrcBot) (string, string, bool) {
	command, ok := args["command"]
	if !ok {
		irc.Reply(m, "Somehow we got the wrong number of arguments.")
		return "", "", false
	}
	identifier, ok := args["nick_or_chan"]
	if !ok {
		irc.Reply(m, "Somehow we got the wrong number of arguments.")
		return "", "", false
	}
	return command, identifier, true
}

// IRC actions
func addACL(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	command, identifier, ok := porcessAclParams(args, m, irc)
	if !ok {
		return false
	}
	// First let's check if the ACL is already present.
	if acl.ExistsACL(command, identifier, irc.DB()) {
		irc.Reply(m, "This ACL is already present.")
		return false
	} else {
		err := acl.SaveACL(command, identifier, irc.DB())
		if err != nil {
			irc.Logger().Error("Problem saving ACLs:", "error", err.Error())
			irc.Reply(m, "Couldn't save the new ACL.")
			return false
		}
	}
	irc.Reply(m, "The ACL was saved.")
	return true
}

// Special command to remove an acl rule
func removeAcl(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	command, identifier, ok := porcessAclParams(args, m, irc)
	if !ok {
		return false
	}
	db := irc.DB()

	// First let's check if the ACL is already present.
	if !acl.ExistsACL(command, identifier, db) {
		irc.Reply(m, "This ACL is not present.")
		return false
	} else {
		err := acl.DeleteACL(command, identifier, db)
		if err != nil {
			irc.Logger().Error("Problem removing ACLs:", "error", err.Error())
			irc.Reply(m, "Couldn't remove the ACL.")
			return false
		}
	}
	irc.Reply(m, "The ACL was succesfully removed.")
	return true
}

func readAcl(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	command := args["command"]
	myAcl, err := acl.GetACL(command, irc.DB(), irc.Config())
	if err != nil {
		irc.Reply(m, "Could not fetch the requested ACL:")
		irc.Reply(m, err.Error())
		return true
	}
	data := myAcl.Dump()
	irc.Reply(m, fmt.Sprintf("ACL for %s", command))
	irc.Reply(m, "Users:")
	for _, nick := range data["nicks"] {
		irc.Reply(m, fmt.Sprintf("\t%s", nick))
	}
	irc.Reply(m, "Channels:")
	for _, channel := range data["channels"] {
		irc.Reply(m, fmt.Sprintf("\t%s", channel))
	}
	return true
}

func changePass(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	newPass := args["new_password"]
	// Make a message to nickserv. I know this is hacky, but better than forging a message from scratch.
	requestor := m.From
	m.From = "NickServ"
	irc.Reply(m, fmt.Sprintf("SET PASSWORD %s", newPass))
	m.From = requestor
	irc.Reply(m, "Password changed. Do not forget to change the configuration too.")
	return false
}

func part(args map[string]string, m *hbot.Message, irc *IrcBot) bool {
	for _, ch := range irc.Config().Channels {
		irc.bot.Irc.Part(ch, "leaving.")
	}
	go func() {
		// Amazingly, this doesn't work.
		irc.bot.Irc.Close()
	}()
	return true
}
