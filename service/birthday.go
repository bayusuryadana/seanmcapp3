package service

import (
	"fmt"
	"log"
	"seanmcapp/external"
	"seanmcapp/repository"
	"strconv"
	"time"
)

type BirthdayService interface {
	Run()
}

type BirthdayServiceImpl struct {
	PeopleRepo     repository.PeopleRepo
	TelegramClient external.TelegramClient
	ChatId         int64
}

func (b *BirthdayServiceImpl) Run() {
	now := time.Now()
	tmr := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)

	today := b.sendForDay(now.Day(), int(now.Month()), 0)
	tomorrow := b.sendForDay(tmr.Day(), int(tmr.Month()), 1)
	nextWeekPpl := b.sendForDay(nextWeek.Day(), int(nextWeek.Month()), 7)

	fmt.Println(strconv.Itoa(today+tomorrow+nextWeekPpl) + " people has birthday today")
}

func (b *BirthdayServiceImpl) sendForDay(day, month, numOfDays int) int {
	people, err := b.PeopleRepo.Get(day, month)
	if err != nil {
		log.Println("Error fetching birthdays:", err)
		return 0
	}

	for _, p := range people {
		b.sendBirthdayReminder(p, numOfDays)
	}

	return len(people)
}

func (b *BirthdayServiceImpl) sendBirthdayReminder(p repository.People, numOfDays int) {
	dayWords := map[int]string{
		0: "Today",
		1: "Tomorrow",
		7: "Next week",
	}

	dayWord, ok := dayWords[numOfDays]
	if !ok {
		log.Printf("Invalid numOfDays: %d", numOfDays)
		return
	}

	msg := fmt.Sprintf("%s is %s's birthday!!", dayWord, p.Name)
	resp, err := b.TelegramClient.SendMessage(b.ChatId, msg)
	if err != nil {
		log.Println("Failed to send message:", err)
		return
	}
	log.Printf("Sent birthday message: %v", resp)
}
