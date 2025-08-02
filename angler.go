package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const SecondsToDays int64 = 60 * 60 * 24

func GetTodayAngleIssue() int {
	// Angle Issue and timestamp at 7am GMT-5 on July 1 2025
	referenceIssue := 1106
	referenceUnixTime := 1751371200

	currentUnixTime := time.Now().Unix()
	timeDiff := (currentUnixTime - int64(referenceUnixTime)) / SecondsToDays
	currentAngleIssue := referenceIssue + int(timeDiff)
	return currentAngleIssue
}

func GetCurrentSeason() int {
	a := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	b := time.Now().UTC()

	if a.After(b) {
		return 0
	}

	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}

	if month < 0 {
		month += 12
		year--
	}

	season := year*12 + month + 1
	return season
}

func GetFailQuotesListMessage(guildId string) string {
	failQuotes := ListFailQuotes(guildId)
	message := ""
	for i, failQuote := range failQuotes {
		message += fmt.Sprintf("%d. %s\n", i+1, failQuote.Quote)
	}
	return message
}

func GetFailQuoteActionResultMessage(message string, guildId string) string {
	command := strings.Split(message, " ")

	if len(command) == 1 {
		return `!failquotes handles the list of quotes shown when failing to guess the angle
-------------
!failquotes list - Show the current list of quotes
!failquotes add "quote" - Add a new quote to the list
!failquotes remove "position" - Remove quote at the given position`
	}

	action := command[1]
	input := ""

	if len(command) > 2 {
		input = command[2]
	}

	if action == "list" {
		return GetFailQuotesListMessage(guildId)
	} else if input == "" {
		return fmt.Sprintf("Action %s requires an input parameter", action)
	}

	if action == "add" {
		fullQuote := strings.Join(command[2:], " ")
		InsertFailQuote(guildId, fullQuote)
		message := fmt.Sprintf("Added quote '%s'\n", fullQuote)
		message += GetFailQuotesListMessage(guildId)
		return message
	}

	if action == "remove" {
		failQuotes := ListFailQuotes(guildId)
		pos, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Sprintf("Cant convert %s to an integer", input)
		}

		if pos < 1 {
			return "The position of the quote to remove should be greater than 0"
		}

		if pos-1 > len(failQuotes) {
			return "The position is greater than the total lenght of quotes"
		}

		failQuoteId := failQuotes[pos-1].Id
		err = RemoveFailQuote(failQuoteId, guildId)
		if err != nil {
			return err.Error()
		}
		message := fmt.Sprintf("Removed quote '%s' from the list\n", failQuotes[pos-1].Quote)
		message += GetFailQuotesListMessage(guildId)
		return message
	}

	return ""
}

func GetStatsMessage(message string, userId string) string {
	allSeasons := false
	season := GetCurrentSeason()
	command := strings.Split(message, " ")
	var statsMessage string
	var err error
	if len(command) > 1 {
		seasonStr := command[1]
		if seasonStr == "all" {
			allSeasons = true
		} else {
			season, err = strconv.Atoi(seasonStr)
			if err != nil {
				return fmt.Sprintf("%s is not a valid season!!", seasonStr)
			}
		}
	}

	if season < 1 {
		return "Seasons starts at 1!!"
	} else if season > GetCurrentSeason() {
		return fmt.Sprintf("We are only on season %d", season)
	}

	statsMessage = GetStats(userId, season, allSeasons)
	return statsMessage
}
