package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

// This function creates clients
// TODO add client interface
func createClient(channel chan UserRequest, commands []inputCommand) {
	var waitGroup sync.WaitGroup
	for _, cmd := range commands {
		var responseChannel = make(chan ServerResponse)
		req := UserRequest{Action: ReserveEvent, EventId: *cmd.id, TicketCount: cmd.value, responses: responseChannel}
		if cmd.id == nil {
			req.Action = GetListEvents
		}
		waitGroup.Add(1)
		go sendUserRequest(req, channel, &waitGroup)
	}

	//var responseChannel = make(chan ServerResponse)
	//req := UserRequest{Action: GetListEvents, EventId: event.ID, TicketCount: 5, responses: responseChannel}
	//waitGroup.Add(1)
	//go sendUserRequest(req, channel, &waitGroup)
	//
	//var responseChannel2 = make(chan ServerResponse)
	//req2 := UserRequest{Action: ReserveEvent, EventId: event.ID, TicketCount: 5, responses: responseChannel2}
	//waitGroup.Add(1)
	//go sendUserRequest(req2, channel, &waitGroup)

	waitGroup.Wait()
	close(channel)
}

type inputCommand struct {
	id    *string
	value int
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

func userInterface() []inputCommand {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a character to start read from file: ")
	_, _ = reader.ReadString('\n')

	file, err := os.Open("input.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	commands := make([]inputCommand, 0)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 1 {
			num, err := strconv.Atoi(fields[0])
			if err != nil {
				fmt.Println("Invalid command:", line)
				return nil
			}
			commands = append(commands, inputCommand{value: num})
		} else if len(fields) == 2 {
			str := fields[0]
			num, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("Invalid command:", line)
				return nil
			}
			commands = append(commands, inputCommand{id: &str, value: num})
		}
	}
	return commands
}

// creates ticket service and initializes the events of it
func initTicketService() *TicketService {
	var ticketService = new(TicketService)
	ticketService.activeEvents.eventsList = make(map[string]Event)
	return ticketService
}

func main() {
	log.Println("Program Started !")
	var waitGroup sync.WaitGroup
	var ticketService = initTicketService()
	var _ = createEvents(ticketService)
	channel := make(chan UserRequest)
	createServer(&waitGroup, channel, ticketService)
	//waitGroup.Wait()
	log.Printf("Event list At The start: %+v \n\n\n", ticketService.activeEvents.eventsList)
	commands := userInterface()
	if commands != nil {
		//fmt.Println("Parsed commands:")
		//for i, cmd := range commands {
		//	if cmd.id != nil {
		//		fmt.Printf("%d: eventID: %s, ticketCount: %d\n", i+1, *cmd.id, cmd.command)
		//	} else {
		//		fmt.Printf("%d: Command: %d\n", i+1, cmd.command)
		//	}
		//}
		createClient(channel, commands)

	}

	log.Printf("Event list At The End: %+v\n", ticketService.activeEvents.eventsList)
	log.Println("Program Finished!")
}
