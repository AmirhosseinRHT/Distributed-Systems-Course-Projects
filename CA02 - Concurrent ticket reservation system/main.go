package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
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
	ReservedTickets  []string // Keep ID list of reserved tickets
}

type Ticket struct {
	ID      string
	EventId string
}

// Object to send Resesrve Request and get event list through channel
type UserRequest struct {
	Action      int // Action = 0 to get list of events and 1 to reserve
	EventId     string
	TicketCount int
	responses   chan ServerResponse
}

type ServerResponse struct {
	message   string
	eventList []Event
}

// Struct that Maps event to their ID for searching
type EventList struct {
	eventsList map[string]Event
	count      int
}

type TicketService struct {
	activeEvents EventList
	eventCache   sync.Map           // this is for caching and ensures cache is concuren-safe
	cacheSingle  singleflight.Group // this is used for caching and ensures only one go routine populates cache at a time
}

// use google uuid to genereate UUID
func generateUUID() string {
	return uuid.New().String()
}

// Store Event in eventList
func (e *EventList) Store(id string, event *Event) error {
	_, ok := e.eventsList[id]
	if !ok {
		e.eventsList[id] = *event
		e.count++
		log.Println("Stored New Event:", event)
		log.Println("Count events:", e.count)
		return nil
	} else { // TODO do we need to handle the else???? if id already exists?
		log.Println("Event id already Exists!") // TODO: Does this ever happen?
		return fmt.Errorf("Event id already Exists!")
	}
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
	ok := ts.activeEvents.Store(event.ID, event)
	// Here upates the cache
	if ok == nil {
		cachedValue, _ := ts.eventCache.Load("eventList")
		cachedList, _ := cachedValue.([]Event)
		cachedList = append(cachedList, *event)
		ts.eventCache.Store("eventList", cachedList)
		log.Println("Cache Appended for event Name:", event.Name)
	}
	return event, ok
}

func (ts *TicketService) ListEvents() []Event {
	value, ok, _ := ts.cacheSingle.Do("eventList", func() (interface{}, error) {
		var events []Event
		for _, val := range ts.activeEvents.eventsList {
			events = append(events, val)
		}
		return events, nil
	})

	if ok == nil {
		return nil
	}

	cachedList, _ := value.([]Event)
	return cachedList
}

// Decrease available tickets of an event
func (el *EventList) decreaseAvailableTicket(id string, count int) {
	temp := el.eventsList[id]
	temp.AvailableTickets -= count
	el.eventsList[id] = temp
	log.Println("Decresed available tickets for ID: ", id, " count=", count, " available tickets after =", el.eventsList[id].AvailableTickets)
}

func (ts *TicketService) BookTickets(eventID string, numTickets int) ([]string, error) {
	// Implement concurrency control here (Step 3)
	ev, ok := ts.activeEvents.Load(eventID)
	if !ok {
		return nil, fmt.Errorf("event not found")
	}

	if ev.AvailableTickets < numTickets { // TODO: Possible race condition scenario
		return nil, fmt.Errorf("not enough tickets available")
	}

	var ticketIDs []string
	for i := 0; i < numTickets; i++ {
		ticketID := generateUUID()
		ticketIDs = append(ticketIDs, ticketID)
		ev.ReservedTickets = append(ev.ReservedTickets, ticketID)
	}
	ts.activeEvents.decreaseAvailableTicket(eventID, numTickets)
	// ts.activeEvents.Store(eventID, &ev)

	// This part Updates cache if change occured
	cachedValue, _ := ts.eventCache.Load("eventList")
	cachedList, _ := cachedValue.([]Event)
	for i, event := range cachedList {
		if event.ID == eventID {
			event.AvailableTickets -= numTickets
			cachedList[i] = event
			log.Println("Cache updated for eventID:", eventID)
			break
		}
	}
	ts.eventCache.Store("eventList", cachedList)

	return ticketIDs, nil
}

// for each client request we create a thread and run this function that handles the request and sends the response
func (ts *TicketService) handleReceiveUserRequest(req UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	var serverResponse ServerResponse
	if req.Action == GetListEvents { // TODO: Implement caching mechanism
		log.Println("Got a GetListEvent request!")
		serverResponse.message = "List of available events"
		cachedValue, _ := ts.eventCache.Load("eventList")
		serverResponse.eventList, _ = cachedValue.([]Event)
		log.Println("Prepared list of events")
	} else {
		log.Println("Got Reserve ticket request for event: ", req.EventId, " count: ", req.TicketCount)
		tickets, err := ts.BookTickets(req.EventId, req.TicketCount)
		if err != nil {
			serverResponse.message = err.Error()
			log.Println(err.Error())
		} else {
			serverResponse.message = "Reserved Event " + req.EventId + " count: " + strconv.Itoa(req.TicketCount) + " succesfuly"
			log.Println("tickets Reserved with UUIDS: ", tickets, " errors:", err)
		}
	}
	req.responses <- serverResponse
	close(req.responses)
}

// server
func (ts *TicketService) receiveUserRequest(requestChannel <-chan UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	var waitGroup sync.WaitGroup
	for req := range requestChannel {
		waitGroup.Add(1)
		go ts.handleReceiveUserRequest(req, &waitGroup)
	}
	waitGroup.Wait()
}

// in this function we can create one or mutliple server codes to handle clients
func createServer(wg *sync.WaitGroup, channel chan UserRequest, ticketService *TicketService) {
	wg.Add(1)
	go ticketService.receiveUserRequest(channel, wg)
}

// client
func sendUserRequest(req UserRequest, requestChannel chan<- UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	requestChannel <- req
	for req := range req.responses {
		log.Printf("Got Response from server: %+v", req)
	}
}

// This function creates clinets
// TODO add client interface
func createClinet(channel chan UserRequest, event *Event) {
	var waitGroup sync.WaitGroup

	var responseChannel = make(chan ServerResponse)
	req := UserRequest{Action: GetListEvents, EventId: event.ID, TicketCount: 5, responses: responseChannel}
	waitGroup.Add(1)
	go sendUserRequest(req, channel, &waitGroup)

	var responseChannel2 = make(chan ServerResponse)
	req2 := UserRequest{Action: ReserveEvent, EventId: event.ID, TicketCount: 5, responses: responseChannel2}
	waitGroup.Add(1)
	go sendUserRequest(req2, channel, &waitGroup)

	waitGroup.Wait()
	close(channel)
}

// In this function we create events that clients will reserve
func createEvents(ticketService *TicketService) *Event {
	event, err := ticketService.CreateEvent("event0", time.Now(), 100)
	if err != nil {
		log.Println(err)
	}
	ticketService.CreateEvent("event1", time.Now(), 100)
	ticketService.CreateEvent("event2", time.Now(), 100)
	return event // TODO: there would be no return
}

// creates ticket service and initializes the eventlist of it
func initTicketService() *TicketService {
	var ticketService = new(TicketService)
	ticketService.activeEvents.eventsList = make(map[string]Event)
	return ticketService
}

func main() {
	log.Println("Program Started !")
	var waitGroup sync.WaitGroup
	var ticketService = initTicketService()
	var event = createEvents(ticketService) // TODO must avoid return value here

	channel := make(chan UserRequest)

	createServer(&waitGroup, channel, ticketService)
	createClinet(channel, event) // TODO Must not use event as argument

	waitGroup.Wait()
	log.Printf("Event list At The End: %+v", ticketService.activeEvents.eventsList)
	log.Println("Program Finished!")
}
