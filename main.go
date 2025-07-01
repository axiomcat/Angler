package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken       = flag.String("token", "", "Bot access token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
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

func main() {
	CreateTable()

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
		// #Angle #1100 X/4
		// ⬆️⬆️⬆️⬇️: 1° off
		// https://www.angle.wtf/
		lines := strings.Split(m.Content, "\n")
		firstLineValues := strings.Split(lines[0], " ")
		angleNumber, _ := strconv.Atoi(firstLineValues[1][1:])
		numberOfTriesStr := firstLineValues[2][0]
		completed := 1
		if numberOfTriesStr == 'X' {
			completed = 0
			numberOfTriesStr = '4'
		}

		numberOfTries, _ := strconv.Atoi(string(numberOfTriesStr))

		secondLineValues := strings.Split(lines[1], " ")
		angleOff := 0
		// Did not complete
		if len(secondLineValues) > 1 {
			angleOffStr := strings.Split(secondLineValues[1], "°")
			angleOff, _ = strconv.Atoi(angleOffStr[0])
		}
		angleEntry := AngleEntry{UserId: m.Author.ID, GlobalName: m.Author.GlobalName, AngleIssue: angleNumber, Tries: numberOfTries, OffBy: angleOff, Completed: completed}
		InsertEntry(angleEntry)
		ShowAllTable()

		fmt.Printf("Inserted: %s,%s,%d,%d,%d,%v", m.Author.ID, m.Author.GlobalName, angleNumber, numberOfTries, angleOff, completed)

	} else if m.Content == "!standings" {
		standings := GetStandings()
		s.ChannelMessageSend(m.ChannelID, standings)
	} else if strings.HasPrefix(m.Content, "!stats") {
		user := m.Author
		if len(m.Mentions) > 0 {
			user = m.Mentions[0]
		}
		stats := GetStats(user.ID)
		s.ChannelMessageSend(m.ChannelID, stats)
	} else if strings.HasPrefix(m.Content, "!transportador") {
		user := m.Author
		if len(m.Mentions) > 0 {
			user = m.Mentions[0]
		}
		oneGuessEntries := CountOneGuessEntries(user.ID, user.GlobalName)
		s.ChannelMessageSend(m.ChannelID, oneGuessEntries)
	}
}
