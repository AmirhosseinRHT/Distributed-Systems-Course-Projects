package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID               string
	Name             string
	Date             time.Time
	TotalTickets     int
	AvailableTickets int
	ReservedTickets  []string // Keep ID list of reserved tickets
}

type Ticket struct {
	ID      string
	EventId string
}

// Object to send Resesrve Request through channel
type ReserveRequest struct {
	EventId     string
	TicketCount int
}

// Struct that Maps event to their ID for searching
type EventList struct {
	eventsList map[string]Event
	count      int
}

type TicketService struct {
	activeEvents EventList
}

// use google uuid to genereate UUID
func generateUUID() string {
	return uuid.New().String()
}

// Store Event in eventList
func (e *EventList) Store(id string, event *Event) {
	_, ok := e.eventsList[id]
	if !ok {
		e.eventsList[id] = *event
		e.count++
	} // TODO do we need to handle the else???? if id already exists?
}

// Find event From event list
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

// Decrease available tickets of an event
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

	// ev := event.(*Event)
	ev := event
	if ev.AvailableTickets < numTickets {
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		ticketID := generateUUID()
		ticketIDs = append(ticketIDs, ticketID)
		ev.ReservedTickets = append(ev.ReservedTickets, ticketID)
	}
	log.Println("ticket count ", event.AvailableTickets)

	ts.activeEvents.decreaseAvailableTicket(eventID, numTickets)
	log.Println("ticket count ", event.AvailableTickets)

	ev.AvailableTickets -= numTickets
	ts.activeEvents.Store(eventID, &ev)
	return ticketIDs, nil
}

// server
func (ts *TicketService) receiveReserveRequest(requestChannel <-chan ReserveRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range requestChannel {
		tickets, err := ts.BookTickets(req.EventId, req.TicketCount)
		log.Println("ticket", tickets, err)
	}
}

// client
func sendReserveRequest(req ReserveRequest, requestChannel chan<- ReserveRequest) {
	requestChannel <- req
	close(requestChannel)

}

// creates ticket service and initializes the eventlist of it
func initticketService() *TicketService {
	var ticketService = new(TicketService)
	ticketService.activeEvents.eventsList = make(map[string]Event)
	return ticketService
}

func createEvents(ticketService *TicketService) *Event {
	event, err := ticketService.CreateEvent("event0", time.Now(), 100)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Created new Event: ", event)
	}
	return event // TODO: there would be no return
}

func createServer(wg *sync.WaitGroup,channel chan ReserveRequest,ticketService *TicketService){
	wg.Add(1)
	go ticketService.receiveReserveRequest(channel, wg)

}

func createClinet(channel chan ReserveRequest,event *Event){ 
	req := ReserveRequest{EventId: event.ID, TicketCount: 5}
	go sendReserveRequest(req, channel)

}

func main() {
	log.Println("Program Started !")
	var waitGroup sync.WaitGroup
	var ticketService = initticketService()
	var event = createEvents(ticketService) // TODO must avoid return value here

	channel := make(chan ReserveRequest, 100)

	createServer(&waitGroup,channel,ticketService)
	createClinet(channel,event) // TODO Must not use event as argument

	waitGroup.Wait()
	log.Println(" sfdfd", ticketService.activeEvents.eventsList)
	log.Println("Program Finished!")
}
