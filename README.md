# A modular discord  bot frameword in go

### Built-in commands:
| Command  | Description  | Req. Op  |
|---|---|---|
| !op \<userid\> \| \<@user\> | add one or more users to the operators  | yes  |
| !deop \<userid\> \| \<@user\> | remove one or more users from the operators  | yes  |
| !ops | print the current config via direct message | yes  |
| !delmsg \<server id\> \<message id\> | delete a message  | no  |
| !quit  | kill the bot  | yes  |

### Command line options
    -t [discord autentication token]
	-o [comma separated string of operators for the bot]
	-o [comma separated string of plugins to load]

The following environment variables can be used instead of the above command line options. Any option given on the command line will override the corresponding environment variable. 

    DISCORDTOKEN
    TBOPS

### Example of a minimal bot
```package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/opcow/disgobot"
)

var (
	dToken = flag.String("t", "", "discord autentication token")
)

func main() {

	flag.Parse() // flags override env good/bad?

	if *dToken == "" {
		fmt.Println("Usage: dist_twit -t <auth_token>")
		return
	}

	err := disgobot.Run(*dToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	signal.Notify(disgobot.SignalChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-disgobot.SignalChan

	// Cleanly close down the Discord session.
	disgobot.Discord.Close()
}```

