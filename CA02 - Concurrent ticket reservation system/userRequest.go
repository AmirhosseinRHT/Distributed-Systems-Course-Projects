package main

type UserRequest struct {
	Action      int
	EventId     string
	TicketCount int
	responses   chan ServerResponse
	turn        int
}

type ServerResponse struct {
	message   string
	eventList []*Event
}
