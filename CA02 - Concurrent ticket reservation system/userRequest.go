package main

type UserRequest struct {
	Action      int // Action = 0 to get list of events and 1 to reserve
	EventId     string
	TicketCount int
	responses   chan ServerResponse
	turn        int
}

type ServerResponse struct {
	message   string
	eventList []Event
}
