package acl

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lavagetto/ircbot/bot"

	hbot "github.com/whyrusleeping/hellabot"
)

/*
	ACLs management.
*/
type commandACL struct {
	nicks    map[string]bool
	channels map[string]bool
}

func (acl *commandACL) IsAllowed(m *hbot.Message) bool {
	// First check the nickname
	if _, ok := acl.nicks[m.Name]; ok {
		return true
	}
	// Then the channel
	if _, ok := acl.channels[m.To]; ok {
		return true
	}
	return false
}

// CRD operations on ACLs
// GetACL returns a full commandACL that can be used in a command.
func GetACL(ID string, db *sql.DB, conf *bot.Configuration) (*commandACL, error) {
	var c commandACL
	// Admins are always allowed to perform any action.
	c.nicks = make(map[string]bool, 0)
	for _, admin := range conf.Admins {
		c.nicks[admin] = true
	}
	c.channels = make(map[string]bool, 0)
	statement, err := db.Prepare("SELECT identifier FROM acls WHERE command = ?")
	if err != nil {
		return &c, err
	}
	rows, err := statement.Query(ID)
	if err != nil {
		return &c, err
	}
	defer rows.Close()
	for rows.Next() {
		var identifier string
		err := rows.Scan(&identifier)
		if err != nil {
			return &c, err
		}
		if strings.HasPrefix(identifier, "#") {
			c.channels[identifier] = true
		} else {
			c.nicks[identifier] = true
		}
	}
	return &c, err
}

func (c *commandACL) Dump() map[string][]string {
	returnValue := make(map[string][]string, 2)
	returnValue["channels"] = make([]string, len(c.channels))
	i := 0
	for ch := range c.channels {
		returnValue["channels"][i] = ch
		i++
	}
	returnValue["nicks"] = make([]string, len(c.nicks))
	i = 0
	for n := range c.nicks {
		returnValue["nicks"][i] = n
		i++
	}
	return returnValue
}

func ExistsACL(command string, identifier string, db *sql.DB) bool {
	statement, err := db.Prepare("SELECT count(1)  FROM acls WHERE command = ? AND identifier = ?")
	if err != nil {
		return false
	}
	var isPresent int
	err = statement.QueryRow(command, identifier).Scan(&isPresent)
	return err == nil && isPresent == 1
}

func SaveACL(command string, identifier string, db *sql.DB) error {
	statement, err := db.Prepare("INSERT INTO acls VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("could not prepare the statement to add ACLs: %s", err)
	}
	_, err = statement.Exec(command, identifier)
	return err
}

func DeleteACL(command string, identifier string, db *sql.DB) error {
	statement, err := db.Prepare("DELETE FROM acls WHERE command = ? AND identifier = ?")
	if err != nil {
		return fmt.Errorf("could not prepare the statement to remove the  ACL: %s", err)
	}
	_, err = statement.Exec(command, identifier)
	return err
}
