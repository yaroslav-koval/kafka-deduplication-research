package datetime

import "time"

// Age calculates person's age based on birthdate and current time.
func Age(birthDate time.Time) int {
	return AgeFrom(birthDate, time.Now())
}

// AgeFrom calculates person's age based on birthdate and given time.
// Converts compared dates to UTC and clears time part before comparison.
func AgeFrom(birthDate, from time.Time) int {
	birthDate = clearTime(birthDate.UTC())
	from = clearTime(from.UTC())
	years := from.Year() - birthDate.Year()
	birthDay := getAdjustedBirthDay(birthDate, from)

	if from.YearDay() < birthDay {
		years--
	}

	return years
}

func getAdjustedBirthDay(birthDate, from time.Time) int {
	birthDay := birthDate.YearDay()
	currentDay := from.YearDay()

	if isLeap(birthDate) && !isLeap(from) && birthDay >= 60 {
		return birthDay - 1
	}

	if isLeap(from) && !isLeap(birthDate) && currentDay >= 60 {
		return birthDay + 1
	}

	return birthDay
}

func isLeap(date time.Time) bool {
	year := date.Year()
	if year%400 == 0 {
		return true
	} else if year%100 == 0 {
		return false
	} else if year%4 == 0 {
		return true
	}

	return false
}

func clearTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
