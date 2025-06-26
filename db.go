package main

import (
	"database/sql"
	"fmt"
	"log"
	"slices"

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
		User  string
		Score int
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
			scores[userId] = val
		} else {
			newScore := Score{User: globalName, Score: entryScore}
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

	for pos, score := range scoreStandings {
		scoreStr += fmt.Sprintf("%d. %s %d\n", pos+1, score.User, score.Score)
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
			streakCount = 0
		}
		lastIssue = entry.Issue
	}

	fmt.Println(currentStreakFound, currentStreak, streakCount)

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
