package disgobot

import (
	"errors"
	"fmt"
	"os"
	"plugin"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type disgoBot interface {
	BotInit([]string) error
	BotExit()
	MessageProc(*discordgo.MessageCreate, []string) bool
}

// MessageProc is a discord message processor function
type MessageProc func(*discordgo.MessageCreate, []string) bool

var (
	// Discord is the discord session pointer
	Discord      *discordgo.Session
	messageProcs []MessageProc
	botOps       = make(map[string]struct{})
	addS         = map[bool]string{
		false: "",
		true:  "s",
	}
	// SignalChan - send signal for killing the bot
	SignalChan = make(chan os.Signal, 1)
)

// Run starts the bot running.
// Pass the discord bot token as token.
func Run(token string) error {
	var err error
	Discord, err = discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return err
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	Discord.AddHandler(messageCreate)
	// Open a websocket connection to Discord and begin listening.
	err = Discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	return err
}

// LoadPlugin takes a string holding to a path to bot plugin to load.
// Plugins add functionality to the bots basic functions.
// Args for the plugin can be included using a '?' to separate them.
func LoadPlugin(p string) error {
	plugOpts := strings.Split(p, "?")
	plug, err := plugin.Open(plugOpts[0])
	if err != nil {
		return err
	}
	symBot, err := plug.Lookup("Bot")
	if err != nil {
		return err
	}
	var bot disgoBot
	bot, ok := symBot.(disgoBot)
	if !ok {
		return errors.New("unexpected type from module symbol")
	}
	err = bot.BotInit(plugOpts)
	if err == nil {
		messageProcs = append(messageProcs, bot.MessageProc)
	}
	return err
}

// addMessageProc is called by plugins to add their message processing function.
// A plugin should call this in its BotInit() function.
// func addMessageProc(p MessageProc) {
// 	messageProcs = append(messageProcs, p)
// }

// RemMessageProc is called by plugins to remove their message processing function.
// func RemMessageProc(id string) bool {
// 	_, ok := messageProcs[id]
// 	if ok {
// 		delete(messageProcs, id)
// 	}
// 	return ok
// }

// IsOp is passed a user ID and will return true if the user is a bot operator.
func IsOp(id string) bool {
	if _, ok := botOps[id]; ok {
		return true
	}
	return false
}

// AddOp adds a bot operator.
func AddOp(id string) {
	botOps[id] = struct{}{}
}

// RemOp removes a bot operator.
func RemOp(id string) {
	delete(botOps, id)
}

func opUsers(users []*discordgo.User, deop bool) int {
	count := 0
	// users := m.Mentions
	for _, u := range users {
		_, ok := botOps[u.ID]
		if ok && deop {
			delete(botOps, u.ID)
			count++
		} else if !ok && !deop {
			botOps[u.ID] = struct{}{}
			count++
		}
	}
	return count
}

func idsToUsers(ids []string) []*discordgo.User {
	var users []*discordgo.User
	for _, id := range ids {
		u, err := Discord.User(id)
		if err == nil {
			users = append(users, u)
		}
	}
	return users
}

// UserIDtoMention takes a user ID and returns a string formatted as a discord mention.
func UserIDtoMention(id string) string {
	u, err := Discord.User(id)
	if err == nil {
		return u.Mention()
	}
	return id
}

// ChanIDtoMention takes a channel ID and returns a discord channel mention.
// If it fails it returns "channel: " + the channel ID string.
func ChanIDtoMention(id string) string {
	channel, err := Discord.State.Channel(id)
	if err == nil {
		return channel.Mention()
	}
	return "channel: " + id
}

// ChanMentionToID takes a channel mention and returns a discord channel ID.
// If passed a valid ID it is returns it unchanged.
func ChanMentionToID(mention string) (id string, err error) {
	id = strings.Replace(strings.Replace(strings.Replace(mention, "<", "", 1), ">", "", 1), "#", "", 1)
	_, err = Discord.Channel(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func showOps(id string) {
	if IsOp(id) {
		c, err := Discord.UserChannelCreate(id)
		if err != nil {
			return
		}
		s := "operators:"
		for k := range botOps {
			s = s + " " + UserIDtoMention(k)
		}
		Discord.ChannelMessageSend(c.ID, s)
	}
}

func denyCommand(id string) {
	c, err := Discord.UserChannelCreate(id)
	if err == nil {
		Discord.ChannelMessageSend(c.ID, "You are not an operator of this bot.")
	}
}

func checkOpAndDeny(id string) bool {
	var status bool
	if status = IsOp(id); !status {
		denyCommand(id)
	}
	return status
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	msg := strings.Split(m.Content, " ")
	for i, f := range messageProcs {
		if f != nil && !f(m, msg) {
			// if the message proc returns false stop calling it
			messageProcs[i] = nil
		}
	}
	switch msg[0] {
	case "!op":
		if !checkOpAndDeny(m.Author.ID) {
			return
		}
		count := 0
		users := idsToUsers(msg[1:])
		if len(users) > 0 {
			count += opUsers(users, false)
		}
		count += opUsers(m.Mentions, false)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d user%s added to operators.", count, addS[count != 1]))
	case "!deop":
		if !checkOpAndDeny(m.Author.ID) {
			return
		}
		count := 0
		users := idsToUsers(msg[1:])
		if len(users) > 0 {
			count += opUsers(users, true)
		}
		count += opUsers(m.Mentions, true)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%d user%s removed from operators.", count, addS[count != 1]))
	case "!delmsg":
		if !checkOpAndDeny(m.Author.ID) {
			return
		}
		if len(msg) > 2 {
			s.ChannelMessageDelete(msg[1], msg[2])
		}
	case "!ops":
		showOps(m.Author.ID)
	case "!quit":
		if m.Message.GuildID == "" {
			if !checkOpAndDeny(m.Author.ID) {
				return
			}
			Discord.Close()
			fmt.Println("Quitting.")
			SignalChan <- os.Kill
		}
	}
}
