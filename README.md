# ircbot
Simple IRC bot written in golang. Mostly a toy.

It can be built as any common go application.

## Running ircbot

It accepts a single command-line parameter, `-config`, allowing to pass the name of the config file to read (which is `config.json` by default).

A typical configuration file will look as follows:

```json
{
    "server": "irc.mynetwork.com",
    "port": 6697,
    "use_tls": true,
    "nick": "IrcbotBot",
    "password": "mysecretpassword",
    "use_sasl": true,
    "channels": ["#channel1", "#channel2"],
    "db_dsn": "sqlite:///srv/ircbot/ircbot.db
}
```
To generate the schema of the database, run:
```bash
sqlite3 ircbot.db < schema.sql
```

## Available Commands.

You can list the implemented commands using `!help`

## ACLs

Only people listed as admins in the configuration will have free access to all commands.

You can grant one user, or a channel the right to use a command as follows:

```
# Allow a user to use a command
you > !acl_add contact_add SomeFriend
IrcbotBot>	The ACL was saved.
# See the acl
you > !acl_get contact_add
IrcbotBot>	ACL for contact_add
IrcbotBot>	Users:
IrcbotBot>		you
IrcbotBot>		SomeFriend
IrcbotBot>	Channels:
# Allow all users in a channel to use a command
you > !acl_add contact_add #thischan
IrcbotBot>	The ACL was saved.
# Remove the authorization to a user
you > !acl_remove contact_add SomeFriend
IrcbotBot>	The ACL was succesfully removed.
```

## How to use the bot
You just need to initialize it in your main program
```golang
    irc, err := ircbot.Init("config.json")
    if err != nil {
        panic(err)
    }
    // Add your commands here
    //
    // Now run it
    irc.Run()
```

## How to add commands

As a reference, commands to manage a contact list are provided under example/

Basically, you have to pick a name for the command, and a callback to be called from it.

The signature of this callback needs to be `ircbot.CommandAction`.
So for example:
```golang
    func sayHello(args []string, m *hbot.Message, i *ircbot.IrcBot) bool {
        greeted := m.From
        if args[0] != "" {
            greeted = args[0]
        }
        i.Reply(m, fmt.Sprintf("Hello, %s!", greeted))
        return true
    }

    hello := irc.AddCommand("greet", sayHello)
```
now we want to define the arguments that this command accepts. Those will be passed to the callback in the
args slice. We also want to set a help message. Use named parameters in the regular expression for the arguments
as it will print out a nicer help message later on. Finally, we want this to work in public channels, so we call
`AllowChannel()`. If we want to allow greeting in query, we'll need to also add `AllowPrivate()`
```golang
    hello.Arguments("(?P<name>\\w+)?")
    hello.SetHelp("Cheers someone").AllowChannel()
```

### A more complex example: a contact list

Very simple interface, you add a new contact with `!contact_add`, and retrieve it with `!contact_get`,
but it shows how to store and retrieve information in the database.


## FAQ
Q: Is ircbot useful for X?
A: No.