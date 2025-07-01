package main

import (
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
