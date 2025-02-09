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

func sendUserRequest(req UserRequest, requestChannel chan<- UserRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	requestChannel <- req
	for req := range req.responses {
		if req.eventList == nil {
			log.Printf("Got Response from server: %s \n", req.message)

		} else {
			log.Printf("Got Response from server: %s \n", req.message)
			for i, e := range req.eventList {
				fmt.Printf("Event %d : %s  , %s  , %s ,  %d  \n", i, e.ID, e.Name, e.Date.String(), e.AvailableTickets)
			}

		}
	}
}

func createEvents(ticketService *TicketService) {
	ticketService.CreateEvent("event0", time.Now(), 100)
	ticketService.CreateEvent("event1", time.Now(), 100)
	ticketService.CreateEvent("event2", time.Now(), 100)
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
			commands = append(commands, inputCommand{value: num, id: nil})
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

func initTicketService() *TicketService {
	var ticketService = new(TicketService)
	ticketService.activeEvents.eventsList = make(map[string]*Event)
	ticketService.reservedTickets = make(map[string]string)
	return ticketService
}

func main() {
	log.Println("Program Started !")
	var waitGroup sync.WaitGroup
	var ticketService = initTicketService()
	createEvents(ticketService)
	channel := make(chan UserRequest) // channel to connect client and server
	createServer(&waitGroup, channel, ticketService)
	commands := userInterface()
	if commands != nil {
		ticketService.createClient(channel, commands)
	}
	for event := range ticketService.activeEvents.eventsList { // logging status of events at the end of program
		log.Printf("Event list At The End: %+v\n", ticketService.activeEvents.eventsList[event])
	}
	log.Println("Program Finished!")
}
