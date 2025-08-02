package main

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strconv"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var tableStatements = map[string]string{
	"angle_tries": "CREATE table angle_tries(id text not null primary key, user_id text, global_name text, angle_issue integer, tries integer, off_by integer, completed integer);",
	"fail_quotes": "CREATE table fail_quotes(id text not null primary key, guild_id text, quote text);"}

type AngleEntry struct {
	UserId     string
	GlobalName string
	AngleIssue int
	Tries      int
	OffBy      int
	Completed  int
}

type QuoteEntry struct {
	Id      string
	Quote   string
	GuildId string
}

func checkTableExist(tableName string, db *sql.DB) bool {
	stmt, err := db.Prepare("SELECT name FROM sqlite_master WHERE type='table' AND name=?;")
	if err != nil {
		log.Fatal("Error in statment", err)
	}
	defer stmt.Close()

	tableExists := ""
	err = stmt.QueryRow(tableName).Scan(&tableExists)
	if err != nil {
		log.Println("Error in query", err)
		return false
	}
	return tableName == tableExists
}

func CreateTables() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for tableName, tableStament := range tableStatements {
		tableExist := checkTableExist(tableName, db)
		if !tableExist {
			fmt.Println("Creating table", tableName)
			_, err = db.Exec(tableStament)
			if err != nil {
				log.Printf("%q: %s\n", err, tableStament)
				return
			}
		}
	}
}

func InsertAngleTryEntry(angleEntry AngleEntry) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into angle_tries(id, user_id, global_name, angle_issue, tries, off_by, completed, season) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	currentSeason := GetCurrentSeason()

	newId := uuid.NewString()
	_, err = stmt.Exec(newId, angleEntry.UserId, angleEntry.GlobalName, angleEntry.AngleIssue, angleEntry.Tries, angleEntry.OffBy, angleEntry.Completed, currentSeason)
	if err != nil {
		log.Println("error in stmt exc")
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("error in stmt commit")
		log.Fatal(err)
	}
}

func ShowAngleTriesTable() {
	db, err := sql.Open("sqlite3", "./foo.db")
	rows, err := db.Query("select * from angle_tries")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var userId string
		var globalName string
		var angleIssue int
		var tries int
		var offBy int
		var completed int
		var season int

		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed, &season)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, userId, globalName, angleIssue, tries, offBy, completed, season)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func calculateEntryScore(tries int, completed int) int {
	if completed == 0 {
		return 5
	}

	if tries == 1 {
		return 100
	} else if tries == 2 {
		return 50
	} else if tries == 3 {
		return 30
	} else {
		return 15
	}
}

func GetStandings() string {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("select * from angle_tries where season = ?")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	currentSeason := GetCurrentSeason()
	rows, err := stmt.Query(currentSeason)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type Score struct {
		User   string
		Score  int
		Played int
		Wins   int
	}

	scores := map[string]Score{}

	for rows.Next() {
		var id string
		var userId string
		var globalName string
		var angleIssue int
		var tries int
		var offBy int
		var completed int
		var season int

		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed, &season)
		if err != nil {
			log.Fatal(err)
		}

		entryScore := calculateEntryScore(tries, completed)

		if val, ok := scores[userId]; ok {
			val.Score += entryScore
			val.Played += 1
			if completed == 1 {
				val.Wins += 1
			}
			scores[userId] = val
		} else {
			newScore := Score{User: globalName, Score: entryScore, Played: 1, Wins: 0}
			if completed == 1 {
				newScore.Wins += 1
			}
			scores[userId] = newScore
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	scoreStandings := []Score{}

	for _, v := range scores {
		scoreStandings = append(scoreStandings, v)
	}

	slices.SortFunc(scoreStandings, func(a Score, b Score) int {
		if a.Score > b.Score {
			return -1
		}
		return 1
	})

	scoreStr := fmt.Sprintf("Season %d\n", currentSeason)

	longestUsername := 0
	longestScore := 0

	for _, score := range scoreStandings {
		nameLen := len(score.User)
		longestUsername = max(nameLen, longestUsername)
		scoreLen := len(strconv.Itoa(score.Score))
		longestScore = max(longestScore, scoreLen)
	}

	for pos, score := range scoreStandings {
		percentageWin := 100.0 * float32(score.Wins) / float32(score.Played)
		scoreStr += fmt.Sprintf("%2d. %*s %*d (%.0f%% win)\n", pos+1, longestUsername, score.User, longestScore, score.Score, percentageWin)
	}

	return scoreStr
}

func GetStats(userId string, season int, allSeasons bool) string {
	db, err := sql.Open("sqlite3", "./foo.db")
	var stmt *sql.Stmt
	var rows *sql.Rows

	if allSeasons == true {
		stmt, err = db.Prepare("select * from angle_tries where user_id = ? order by angle_issue desc")
		if err != nil {
			log.Fatal(err)
		}
		rows, err = stmt.Query(userId)
	} else {
		stmt, err = db.Prepare("select * from angle_tries where user_id = ? and season = ? order by angle_issue desc")
		if err != nil {
			log.Fatal(err)
		}
		rows, err = stmt.Query(userId, season)
	}
	defer stmt.Close()
	defer rows.Close()

	type Entry struct {
		Tries     int
		Issue     int
		Completed int
	}

	entries := []Entry{}

	for rows.Next() {
		var id string
		var userId string
		var globalName string
		var angleIssue int
		var tries int
		var offBy int
		var completed int
		var season int

		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed, &season)
		if err != nil {
			log.Fatal(err)
		}
		entry := Entry{Tries: tries, Completed: completed, Issue: angleIssue}
		entries = append(entries, entry)
	}

	wins := 0
	played := len(entries)
	maxStreak := 0
	currentStreak := 0
	currentStreakFound := false
	streakCount := 0

	if len(entries) == 0 {
		return "No games played yet!"
	}

	firstEntry := entries[0]
	if firstEntry.Completed == 1 {
		wins += 1
		maxStreak = 1
		currentStreak = 1
		streakCount = 1
	} else {
		currentStreakFound = true
	}

	lastIssue := firstEntry.Issue

	for _, entry := range entries[1:] {
		if entry.Completed == 1 {
			wins += 1
			if lastIssue-1 == entry.Issue {
				streakCount += 1
			} else {
				if !currentStreakFound {
					currentStreak = streakCount
					currentStreakFound = true
				}
				streakCount = 1
			}

			maxStreak = max(maxStreak, streakCount)
		} else {
			if !currentStreakFound {
				currentStreak = streakCount
				currentStreakFound = true
			}
			streakCount = 0
		}
		lastIssue = entry.Issue
	}

	if !currentStreakFound {
		currentStreak = streakCount
	}

	winPercentage := 100.0 * float32(wins) / float32(played)

	var seasonText string
	if allSeasons == true {
		seasonText = "All seasons\n"
	} else {
		seasonText = fmt.Sprintf("Season %d\n", season)
	}

	stats := seasonText

	stats += fmt.Sprintf("%.1f Win%%\n", winPercentage)
	stats += fmt.Sprintf("%d Played\n", played)
	stats += fmt.Sprintf("%d Current Streak\n", currentStreak)
	stats += fmt.Sprintf("%d Max Streak\n", maxStreak)

	return stats
}

func CountOneGuessEntries(userId string, userName string) string {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("select count(*) from angle_tries where user_id = ? and tries == 1")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var oneGuessEntries int
	err = stmt.QueryRow(userId).Scan(&oneGuessEntries)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s ha usado el transportador <:emoji_22:1383877615613509715> %d veces\n", userName, oneGuessEntries)
}

func GetUsersIds() []string {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("select distinct user_id from angle_tries")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	usersIds := []string{}
	for rows.Next() {
		var userId string

		err = rows.Scan(&userId)
		if err != nil {
			log.Fatal(err)
		}
		usersIds = append(usersIds, userId)
	}
	return usersIds
}

func GetUserIdsAngleIssueDone(angleIssue int) []string {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("select user_id from angle_tries where angle_issue = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(angleIssue)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	usersIds := []string{}
	for rows.Next() {
		var userId string

		err = rows.Scan(&userId)
		if err != nil {
			log.Fatal(err)
		}
		usersIds = append(usersIds, userId)
	}
	return usersIds
}

func InsertFailQuote(guildId string, failQuote string) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into fail_quotes(id, guild_id, quote) values(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	newId := uuid.NewString()
	_, err = stmt.Exec(newId, guildId, failQuote)
	if err != nil {
		log.Println("error in stmt exc")
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("error in stmt commit")
		log.Fatal(err)
	}

	log.Printf("Inserted %s in guild %s", failQuote, guildId)
}

func RemoveFailQuote(failQuoteId string, guildId string) error {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		return fmt.Errorf("Error opening database")
	}
	defer db.Close()

	stmt := fmt.Sprintf("DELETE FROM fail_quotes WHERE id = '%s'", failQuoteId)
	fmt.Println(stmt)
	_, err = db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("Error while removing fail quote %s", err)
	}
	return nil
}

func ListFailQuotes(guildId string) []QuoteEntry {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("SELECT id, quote FROM fail_quotes WHERE guild_id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(guildId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	failQuotes := []QuoteEntry{}

	for rows.Next() {
		var id string
		var quote string

		err = rows.Scan(&id, &quote)
		if err != nil {
			log.Fatal(err)
		}
		quoteEntry := QuoteEntry{Id: id, Quote: quote}
		failQuotes = append(failQuotes, quoteEntry)
	}
	return failQuotes
}
