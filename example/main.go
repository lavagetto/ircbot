package main

// Simple chatbot started to help the wikimedia SRE team during outages
/*
Copyright (C) 2022  Giuseppe Lavagetto

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
*/

import (
	"flag"
	"fmt"

	"github.com/lavagetto/ircbot/example/contact"
	"github.com/lavagetto/ircbot/ircbot"
	hbot "github.com/whyrusleeping/hellabot"

	_ "github.com/mattn/go-sqlite3"
)

var configFile = flag.String("config", "config.json", "Optional configuration file (JSON)")

// Very basic example. For a more complex one see the contact module
func sayHello(args []string, m *hbot.Message, i *ircbot.IrcBot) bool {
	greeted := m.From
	if args[0] != "" {
		greeted = args[0]
	}
	i.Reply(m, fmt.Sprintf("Hello, %s!", greeted))
	return true
}

func main() {
	flag.Parse()
	irc, err := ircbot.Init(*configFile)
	if err != nil {
		panic(err)
	}
	// Add one command
	irc.AddCommand("greet", sayHello).Arguments("(?P<name>\\w+)?").SetHelp("Cheer the counterpart").AllowChannel()
	// Add commands from the contact list module
	contact.AddContact(irc)
	irc.Run()
}
