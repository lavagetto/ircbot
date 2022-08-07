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

Basically, you have to pick a name for the command, and a callback to be called from it. So say you wanted to make a basic greeter function, that replies to `!greet <name>`:
```golang
func sayHello(args map[string]string, m *hbot.Message, i *ircbot.IrcBot) bool {
    i.Reply(m, fmt.Sprintf("Hello, %s!", args["name"]))
    // We don't want other handlers to process this message
    return true
}

irc.AddCommand("greet", sayHello).AddParameter("name", `\w+`).AllowPublic()
```
as you can see, the signature of this callback needs to be `ircbot.CommandAction`, and the first argument contains the values of the parameters in
a map. We added `AllowPublic()` to allow the command to be called in public
channels, and the corresponding `AllowPrivate()` to allow the command in private.

Please note: if you don't add either, your command will not be invoked in any situation!

The command parser is very strict, and if a parameter is not found, it will
refuse to execute the command.

It is however possible to add a default value for a parameter using `AddParameterWithDefault`, or a context-dependent default using
`AddParameterWithDefaultCb`.

 For instance, let's say we want our "greet" function
to default to the sender name if none was provided.
So for example:
```golang
    func defaultGreeter(m *hbot.Message) string {
        return m.From
    }
    c := irc.AddCommand("greet", sayHello)
    c.AddParameterWithDefaultCb("name", `\w+`, defaultGreeter)
```
 We also want to set a help message. That is done by using the `Help` method.
 Ircbot will take care of properly formatting the output for you, including
 an example of the syntax with parameters.


### A more complex example: a contact list

Very simple interface, you add a new contact with `!contact_add`, and retrieve it with `!contact_get`,
but it shows how to store and retrieve information in the database.


## FAQ
### Is ircbot useful for X?
No. In fact, you should not use it.

### Can you add feature X?
Didn't you see the `patches welcome` sign out front?

### Is ircbot ready for production?
It depends. Do you employ the author? If yes, "maybe".
Otherwise, "LOL".

### Is ircbot cloud-native?
Get lost.

### Isn't IRC like an internet protocol from the 90s?

Yes, so is HTTP. So what?

### Ok by why not a slack bot?

Because I'm an anti-business commie boomer, obviously. Also see "Is ircbot cloud-native?"
