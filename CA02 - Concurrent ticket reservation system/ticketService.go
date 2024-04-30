package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type TicketService struct {
	activeEvents    EventList
	eventCache      sync.Map           // this is for caching and ensures cache is concurrency-safe
	cacheSingle     singleflight.Group // this is used for caching and ensures only one go routine populates cache at a time
	reservedTickets map[string]string
}

func (ts *TicketService) ListEvents() []Event {
	value, ok, _ := ts.cacheSingle.Do("eventList", func() (interface{}, error) {
		var events []*Event
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

func (ev *Event) wait(turn int) bool {
	for true {
		if ev.turn == turn {
			return true
		}
	}
	return true
}

func (ts *TicketService) BookTickets(eventID string, numTickets int, reqTurn int) ([]string, error) {
	ev, ok := ts.activeEvents.Load(eventID)
	if !ok {
		return nil, fmt.Errorf("event not found")
	}
	ev.wait(reqTurn)
	ev.mtx.Lock()
	ev.turn += 1
	if ev.AvailableTickets < numTickets {
		ev.mtx.Unlock()
		return nil, fmt.Errorf("not enough tickets available")
	}
	ts.activeEvents.decreaseAvailableTicket(eventID, numTickets)
	cachedValue, _ := ts.eventCache.Load("eventList")
	cachedList, _ := cachedValue.([]*Event)
	for i, event := range cachedList {
		if event.ID == eventID {
			cachedList[i] = event
			log.Println("Cache updated for eventID:", eventID)
			break
		}
	}
	ts.eventCache.Store("eventList", cachedList)
	ev.mtx.Unlock()
	return ts.createTicketIDs(eventID, numTickets), nil
}

func (ts *TicketService) createTicketIDs(eventID string, count int) []string {
	var ticketIDs []string
	for i := 0; i < count; i++ {
		ticketID := generateUUID()
		ticketIDs = append(ticketIDs, ticketID)
		ts.reservedTickets[ticketID] = eventID
	}
	return ticketIDs
}

func (ts *TicketService) createClient(channel chan UserRequest, commands []inputCommand) {
	var waitGroup sync.WaitGroup
	for _, cmd := range commands {
		var responseChannel = make(chan ServerResponse)
		req := UserRequest{Action: ReserveEvent, EventId: *cmd.id, TicketCount: cmd.value,
			responses: responseChannel, turn: ts.activeEvents.eventsList[*cmd.id].waitedCount}
		ts.activeEvents.eventsList[*cmd.id].waitedCount += 1
		if cmd.id == nil {
			req.Action = GetListEvents
		}
		waitGroup.Add(1)
		go sendUserRequest(req, channel, &waitGroup)
	}
	waitGroup.Wait()
	close(channel)
}

// for each client request we create a thread and run this function that handles the request and sends the response
func (ts *TicketService) handleReceiveUserRequest(req UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	var serverResponse ServerResponse
	if req.Action == GetListEvents {
		log.Println("Got a GetListEvent request!")
		serverResponse.message = "List of available events"
		cachedValue, _ := ts.eventCache.Load("eventList")
		serverResponse.eventList, _ = cachedValue.([]Event)
		log.Println("Prepared list of events")
	} else {
		log.Println("Got Reserve ticket request for event: ", req.EventId, " count: ", req.TicketCount)
		tickets, err := ts.BookTickets(req.EventId, req.TicketCount, req.turn)
		if err != nil {
			serverResponse.message = err.Error()
			log.Println(err.Error())
		} else {
			serverResponse.message = "Reserved Event " + req.EventId + " count: " + strconv.Itoa(req.TicketCount) + " successfully"
			for i, ticket := range tickets {
				log.Println(i+1, " : ", ticket)
			}
		}
	}
	req.responses <- serverResponse
	close(req.responses)
}

func (ts *TicketService) receiveUserRequest(requestChannel <-chan UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	var waitGroup sync.WaitGroup
	for req := range requestChannel {
		waitGroup.Add(1)
		go ts.handleReceiveUserRequest(req, &waitGroup)
		//time.Sleep(time.Millisecond)
	}
	waitGroup.Wait()
}

func (ts *TicketService) CreateEvent(name string, date time.Time, totalTickets int) (*Event, error) {
	event := &Event{
		ID:               strconv.Itoa(ts.activeEvents.count), // Generate a unique ID for the event
		Name:             name,
		Date:             date,
		TotalTickets:     totalTickets,
		AvailableTickets: totalTickets,
		waitedCount:      0,
		turn:             0,
	}
	ok := ts.activeEvents.Store(event.ID, event)
	// Here updates the cache
	if ok == nil {
		cachedValue, _ := ts.eventCache.Load("eventList")
		cachedList, _ := cachedValue.([]*Event)
		cachedList = append(cachedList, event)
		ts.eventCache.Store("eventList", cachedList)
		log.Println("Cache Appended for event Name:", event.Name)
	}
	return event, ok
}
