package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	GetListEvents = 0
	ReserveEvent  = 1
)

type Event struct {
	ID               string
	Name             string
	Date             time.Time
	TotalTickets     int
	AvailableTickets int
	mtx              sync.Mutex
	waitedCount      int
	turn             int
}

type EventList struct {
	eventsList map[string]*Event
	count      int
}

func generateUUID() string {
	return uuid.New().String()
}

// Store Event in eventList
func (e *EventList) Store(id string, event *Event) error {
	_, ok := e.eventsList[id]
	if !ok {
		e.eventsList[id] = event
		e.count++
		log.Println("Stored New Event:", event)
		log.Println("Count events:", e.count)
		return nil
	} else { // TODO do we need to handle the else???? if id already exists?
		log.Println("Event id already Exists!") // TODO: Does this ever happen?
		return fmt.Errorf("event id already Exists")
	}
}

func (e *EventList) Load(id string) (*Event, bool) {
	val, ok := e.eventsList[id]
	return val, ok
}

func (el *EventList) decreaseAvailableTicket(id string, count int) {
	temp := el.eventsList[id]
	temp.AvailableTickets -= count
	el.eventsList[id] = temp
	log.Println("Decreased available tickets for ID: ", id, " count :", count, " available tickets after =", el.eventsList[id].AvailableTickets)
}
