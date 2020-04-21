package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/opcow/discbot"
)

var lastCD time.Time
var start = make(chan int)
var quit = make(chan bool)

var seed = rand.NewSource(time.Now().Unix())
var rnd = rand.New(seed)

var (
	dToken    = flag.String("t", "", "discord autentication token")
	operators = flag.String("o", "", "comma separated string of operators for the bot")
	plugins   = flag.String("p", "", "comma separated string of bot plugins to load")
	onOrOff   = map[bool]string{
		false: "off",
		true:  "on",
	}
	addS = map[bool]string{
		false: "",
		true:  "s",
	}
	// bot operators
	botOps map[string]struct{}
	sc     chan os.Signal
)

func getEnv() {
	*dToken = os.Getenv("DISCORDTOKEN")
	*operators = os.Getenv("TBOPS")
}

func main() {
	getEnv()
	flag.Parse() // flags override env good/bad?

	if *dToken == "" {
		fmt.Println("Usage: dist_twit -t <auth_token>")
		return
	}

	botOps = make(map[string]struct{})
	for _, c := range strings.Split(*operators, ",") {
		botOps[c] = struct{}{}
	}

	err := discbot.Run(*dToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, p := range strings.Split(*plugins, ",") {
		discbot.LoadPlugin(p)
	}
	// discbot.LoadPlugin("../discbotplugin/pongbot.o")
	// discbot.LoadPlugin("../discbotplugin/reactbot.o")
	// discbot.LoadPlugin("../discbotplugin/covidbot.o")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discbot.Discord.Close()
}
