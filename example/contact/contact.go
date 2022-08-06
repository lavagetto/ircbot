package contact

import (
	"database/sql"
	"fmt"

	"github.com/lavagetto/ircbot/ircbot"

	hbot "github.com/whyrusleeping/hellabot"
)

type Contact struct {
	name  string
	phone string
	email string
}

// Get a contact from the db
func GetContact(db *sql.DB, name string) (*Contact, error) {
	var c Contact
	err := db.QueryRow(
		"SELECT name, phone, email FROM contacts WHERE name = ?",
		name).Scan(&c.name, &c.phone, &c.email)
	return &c, err
}

func (ct *Contact) insert(db *sql.DB) error {
	statement, err := db.Prepare("INSERT INTO contacts VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec(ct.name, ct.phone, ct.email)
	return err
}

func (ct *Contact) update(db *sql.DB) error {
	statement, err := db.Prepare("UPDATE contacts SET phone = ?, email = ? WHERE name = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(ct.phone, ct.email, ct.name)
	return err
}

func (ct *Contact) Save(db *sql.DB) error {
	_, err := GetContact(db, ct.name)
	if err != nil {
		return ct.insert(db)
	} else {
		return ct.update(db)
	}
}

func (ct *Contact) PrettyPrint() string {
	return fmt.Sprintf("%s: %s (%s)", ct.name, ct.phone, ct.email)
}

func (ct *Contact) Remove(db *sql.DB) error {
	statement, err := db.Prepare("DELETE FROM contacts WHERE name = ?")
	if err != nil {
		return err
	}
	_, err = statement.Exec(ct.name)
	return err
}

func addContact(args map[string]string, m *hbot.Message, irc *ircbot.IrcBot) bool {
	contact := Contact{name: args["name"], phone: args["intl_phone"], email: args["email"]}
	err := contact.Save(irc.DB())
	if err == nil {
		irc.Reply(m, "Contact added successfully.")
	} else {
		irc.Reply(m, "Trouble saving the contact, please try again later.")
		irc.Logger().Error(err.Error())
	}
	return true
}

func removeContact(args map[string]string, m *hbot.Message, irc *ircbot.IrcBot) bool {
	db := irc.DB()
	log := irc.Logger()
	contact, err := GetContact(db, args["name"])
	if err != nil {
		irc.Reply(m, "Couldn't find the contact you searched for")
		log.Error(err.Error())
		return true
	}
	err = contact.Remove(db)
	if err != nil {
		irc.Reply(m, "Couldn't remove contact, check logs for the error.")
		log.Error("Error removing contact:", "error", err.Error(), "contact", contact.PrettyPrint())
	} else {
		irc.Reply(m, "Contact successfully removed.")
	}
	return true
}

func getContact(args map[string]string, m *hbot.Message, irc *ircbot.IrcBot) bool {
	db := irc.DB()
	log := irc.Logger()
	contact, err := GetContact(db, args["name"])
	if err != nil {
		irc.Reply(m, "Couldn't find the contact you searched for")
		log.Error(err.Error())
	} else if contact.phone == "" {
		irc.Reply(m, "No phone data for the contact")
	} else {
		irc.Reply(m, contact.PrettyPrint())
	}
	return true
}

func AddContact(irc *ircbot.IrcBot) {
	add := irc.AddCommand("contact_add", addContact).SetHelp("Add a contact (privmsg only)")
	add.Arguments("(?P<name>\\w+)\\s+(?P<intl_phone>\\+\\d{5,15})\\s+(?P<email>\\S+)$").AllowPrivate()
	irc.AddCommand("contact_get", getContact).SetHelp("Gets information about a contact (privmsg only)").Arguments("(?P<name>\\w+)").AllowPrivate()
	irc.AddCommand("contact_remove", removeContact).SetHelp("Removes a contact (privmsg only)").Arguments("(?P<name>\\w+)").AllowPrivate()
}
