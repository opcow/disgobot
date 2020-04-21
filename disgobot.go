package disgobot

import (
	"fmt"
	"os"
	"plugin"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type discBot interface {
	BotInit()
	BotExit()
}

var (
	Discord     *discordgo.Session
	messageProc []func(*discordgo.MessageCreate, []string)
	botOps      map[string]struct{}
	addS        = map[bool]string{
		false: "",
		true:  "s",
	}
)

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

func LoadPlugin(p string) {
	plug, err := plugin.Open(p)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	symBot, err := plug.Lookup("DiscBot")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var bot discBot
	bot, ok := symBot.(discBot)
	if !ok {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}
	bot.BotInit()
	// fmt.Println(botFuncs.ChanIDtoMention("foo"))
}

func AddMessageProc(p func(*discordgo.MessageCreate, []string)) {
	messageProc = append(messageProc, p)
}

func IsOp(id string) bool {
	if _, ok := botOps[id]; ok {
		return true
	}
	c, err := Discord.UserChannelCreate(id)
	if err == nil {
		Discord.ChannelMessageSend(c.ID, "You are not an operator of this bot.")
	}
	return false
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

func UserIDtoMention(id string) string {
	u, err := Discord.User(id)
	if err == nil {
		return u.Mention()
	}
	return id
}

// Converts a channel ID to a mention. On error it returns the channel ID string.
func ChanIDtoMention(id string) string {
	channel, err := Discord.State.Channel(id)
	if err == nil {
		return channel.Mention()
	}
	return "channel: " + id
}

// Converts a channel link to an ID. If passed a valid ID it is returned it unchanged.
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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	msg := strings.Split(m.Content, " ")
	for _, f := range messageProc {
		f(m, msg)
	}
	switch msg[0] {
	case "!op":
		if !IsOp(m.Author.ID) {
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
		if !IsOp(m.Author.ID) {
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
		if len(msg) > 2 {
			s.ChannelMessageDelete(msg[1], msg[2])
		}
	case "!config":
		showOps(m.Author.ID)
	case "!quit":
		if IsOp(m.Author.ID) && m.Message.GuildID == "" {
			Discord.Close()
		}
	}
}