package main

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type Event struct {
	ID               string
	Name             string
	Date             time.Time
	TotalTickets     int
	AvailableTickets int
	ReservedTickets  []string
}

type Ticket struct {
	ID      string
	EventId string
}

// ReserveRequest object must be sent to channel
type ReserveRequest struct {
	EventId     string
	TicketCount int
}

type EventList struct {
	eventsList map[string]Event
	count      int
}

type TicketService struct {
	activeEvents EventList
}

func generateUUID() string {
	return uuid.New().String()
}

func (e *EventList) Store(id string, event *Event) {
	_, ok := e.eventsList[id]
	if !ok {
		e.eventsList[id] = *event
		//fmt.Println(e.eventsList[id])
		e.count++
	}
}

func (e *EventList) Load(id string) (Event, bool) {
	val, ok := e.eventsList[id]
	return val, ok
}

func (ts *TicketService) CreateEvent(name string, date time.Time, totalTickets int) (*Event, error) {
	event := &Event{
		ID:               generateUUID(), // Generate a unique ID for the event
		Name:             name,
		Date:             date,
		TotalTickets:     totalTickets,
		AvailableTickets: totalTickets,
	}
	ts.activeEvents.Store(event.ID, event)
	return event, nil
}

func (ts *TicketService) ListEvents() []Event {
	var events []Event
	for _, val := range ts.activeEvents.eventsList {
		events = append(events, val)
	}
	return events
}

func (el *EventList) decreaseAvailableTicket(id string, count int) {
	temp := el.eventsList[id]
	temp.AvailableTickets -= count
	el.eventsList[id] = temp
}

func (ts *TicketService) BookTickets(eventID string, numTickets int) ([]string, error) {
	// Implement concurrency control here (Step 3)
	fmt.Println("id:", eventID)
	event, ok := ts.activeEvents.Load(eventID)
	if !ok {
		return nil, fmt.Errorf("event not found")
	}

	//ev := event.(*Event)
	ev := event
	if ev.AvailableTickets < numTickets {
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		ticketID := generateUUID()
		ticketIDs = append(ticketIDs, ticketID)
		//fmt.Println("ticket count ", event.AvailableTickets, "\n")
		//event.AvailableTickets--
		//fmt.Println("ticket count ", event.AvailableTickets, "\n")
		ev.ReservedTickets = append(ev.ReservedTickets, ticketID)
	}
	fmt.Println("ticket count ", event.AvailableTickets, "\n")

	ts.activeEvents.decreaseAvailableTicket(eventID, numTickets)
	fmt.Println("ticket count ", event.AvailableTickets, "\n")

	ev.AvailableTickets -= numTickets
	//ts.activeEvents.Store(eventID, ev)
	ts.activeEvents.Store(eventID, &ev)
	return ticketIDs, nil
}

func (ts *TicketService) receiveReserveRequest(requestChannel <-chan ReserveRequest) {
	//for {
	req, ok := <-requestChannel
	if !ok {
		//break
	}
	tickets, err := ts.BookTickets(req.EventId, req.TicketCount)
	fmt.Println("ticket", tickets, err)
	//}
}

func sendReserveRequest(req ReserveRequest, requestChannel chan<- ReserveRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	requestChannel <- req
}

func main() {
	var waitGroup sync.WaitGroup
	var handler = new(TicketService)
	handler.activeEvents.eventsList = make(map[string]Event)
	requests := make(chan ReserveRequest, 100)
	event, err := handler.CreateEvent("sirk", time.Now(), 100)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(event)
	}
	req := ReserveRequest{EventId: event.ID, TicketCount: 5}

	waitGroup.Add(1)
	go sendReserveRequest(req, requests, &waitGroup)

	handler.receiveReserveRequest(requests)
	waitGroup.Wait()

	fmt.Println(" sfdfd", handler.activeEvents.eventsList)
	close(requests)
}
