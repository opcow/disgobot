# A modular discord  bot frameword in go

disgobot uses [discordgo](https://pkg.go.dev/github.com/bwmarrin/discordgo?tab=doc) under the hood.
### Built-in commands:
| Command  | Description  | Req. Op  |
|---|---|---|
| !op \<userid\> \| \<@user\> | add one or more users to the operators  | yes  |
| !deop \<userid\> \| \<@user\> | remove one or more users from the operators  | yes  |
| !ops | print the operators list | yes  |
| !delmsg \<server id\> \<message id\> | delete a message  | yes  |
| !quit  | kill the bot  | yes  |
---
### Example of a minimal bot
```
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/opcow/disgobot"
)

var (
	dToken    = flag.String("t", "", "discord autentication token")
	operators = flag.String("o", "", "comma separated string of operators for the bot as discord IDs")
	plugins   = flag.String("p", "", "comma separated string of bot plugins to load")
)

func main() {

	flag.Parse()

	if *dToken == "" {
		fmt.Println("Usage: dist_twit -t <auth_token>")
		return
	}

	err := disgobot.Run(*dToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Add some ops.
	for _, o := range strings.Split(*operators, ",") {
		if o != "" {
			disgobot.AddOp(o)
		}
	}

	// Load plugins.
	for _, p := range strings.Split(*plugins, ",") {
		if p != "" {
			disgobot.LoadPlugin(p)
		}
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	signal.Notify(disgobot.SignalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-disgobot.SignalChan

	// Cleanly close down the Discord session.
	disgobot.Discord.Close()
}
```
---
### Example plugin
```
package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/opcow/disgobot"
)

type discBot string

var DiscBot discBot

func (b discBot) BotInit() {
	// Tell disgobot where to pass messages for processing
	disgobot.AddMessageProc(messageProc)
}

func (b discBot) BotExit() {
}

// messageProc() receives a doscordgo MessageCreate struct and the
// message content split into an array of words
func messageProc(m *discordgo.MessageCreate, msg []string) {
	if strings.ToLower(m.Content) == "ping" {
		disgobot.Discord.ChannelMessageSend(m.ChannelID, "PONG")
	}
}

```
---