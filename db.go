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

type AngleEntry struct {
	UserId     string
	GlobalName string
	AngleIssue int
	Tries      int
	OffBy      int
	Completed  int
}

func CreateTable() {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table angle_tries(id text not null primary key, user_id text, global_name text, angle_issue integer, tries integer, off_by integer, completed integer);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func InsertEntry(angleEntry AngleEntry) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into angle_tries(id, user_id, global_name, angle_issue, tries, off_by, completed) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	newId := uuid.NewString()
	_, err = stmt.Exec(newId, angleEntry.UserId, angleEntry.GlobalName, angleEntry.AngleIssue, angleEntry.Tries, angleEntry.OffBy, angleEntry.Completed)
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

func ShowAllTable() {
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

		// userId globalName angleIssue tries offBy completed
		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, userId, globalName, angleIssue, tries, offBy, completed)
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
	rows, err := db.Query("select * from angle_tries")
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

		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed)
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

	scoreStr := ""

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

func GetStats(userId string) string {
	db, err := sql.Open("sqlite3", "./foo.db")
	stmt, err := db.Prepare("select * from angle_tries where user_id = ? order by angle_issue desc")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userId)
	if err != nil {
		log.Fatal(err)
	}
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

		err = rows.Scan(&id, &userId, &globalName, &angleIssue, &tries, &offBy, &completed)
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
		fmt.Println(entry)
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

	stats := ""

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
