package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

var (
	BotToken  = flag.String("token", "", "Bot access token")
	ChannelId = flag.String("channel", "", "Channel to send reminder")
)

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func sendReminderMessage(channelId string) {
	userIds := GetUsersIds()
	todayAngleIssue := GetTodayAngleIssue()
	usersTodayAngleDone := GetUserIdsAngleIssueDone(todayAngleIssue)
	userIdsMissingleTodayAngle := []string{}
	for _, userId := range userIds {
		if !slices.Contains(usersTodayAngleDone, userId) {
			userIdsMissingleTodayAngle = append(userIdsMissingleTodayAngle, userId)

		}
	}

	message := ""
	if len(userIds) == len(userIdsMissingleTodayAngle) {
		message = "No one has tried guessing today's angle yet!"
	} else if len(userIdsMissingleTodayAngle) > 0 {
		message = "Remember to do today's angle!"
		for _, userId := range userIdsMissingleTodayAngle {
			message += fmt.Sprintf(" <@%s>", userId)
		}
	}

	if len(message) > 0 {
		log.Println("Sending reminder for Issue ", todayAngleIssue)
		s.ChannelMessageSend(channelId, message)
	}
}

func startCronJobs() {
	c := cron.New()
	c.AddFunc("0 1,13,17,21 * * *", func() { sendReminderMessage(*ChannelId) })
	c.Start()
}

func main() {
	startCronJobs()
	CreateTables()

	// Register the messageCreate func as a callback for MessageCreate events.
	s.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	s.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err := s.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	// Cleanly close down the Discord session.
	s.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "#Angle") {
		angleEntry := ParseAngleEntry(m.Content, m.Author.ID, m.Author.GlobalName)
		InsertAngleTryEntry(angleEntry)
		s.MessageReactionAdd(m.ChannelID, m.Message.ID, GetEntryEmojiReaction(angleEntry.Completed, angleEntry.Tries, m))
	} else if strings.HasPrefix(m.Content, "!standings") {
		s.ChannelMessageSend(m.ChannelID, GetStandingMessage(m.Content))
	} else if strings.HasPrefix(m.Content, "!stats") {
		user := m.Author
		if len(m.Mentions) > 0 {
			user = m.Mentions[0]
		}
		s.ChannelMessageSend(m.ChannelID, GetStatsMessage(m.Content, user.ID))
	} else if strings.HasPrefix(m.Content, "!transportador") {
		user := m.Author
		if len(m.Mentions) > 0 {
			user = m.Mentions[0]
		}
		oneGuessEntries := CountOneGuessEntries(user.ID, user.GlobalName)
		s.ChannelMessageSend(m.ChannelID, oneGuessEntries)
	} else if strings.HasPrefix(m.Content, "!failquotes") {
		s.ChannelMessageSend(m.ChannelID, GetFailQuoteActionResultMessage(m.Content, m.GuildID))
	}
}
