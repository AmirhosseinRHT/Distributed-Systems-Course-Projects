package main

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

// Some related functions might be added here
