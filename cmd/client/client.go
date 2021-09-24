package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tcp-serv-test/internal/client"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		log.Fatal("address must be provided")
	}
	log.Println("starting client")
	c := client.New(arguments[1])
	go c.Start()
	waitStopSignal(c)
}

func waitStopSignal(c *client.Client) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	log.Println("stopping client")
	c.Stop()
	log.Println("stopped")
}
