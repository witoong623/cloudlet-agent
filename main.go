package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	applicationServers := createTestURLs()
	reverseServer, _ := NewMultiHostReverseProxyServer(applicationServers)
	go func() {
		reverseServer.StartReverseProxyServer()
	}()

	clientMonitorCancelSignal := make(chan struct{}, 1)
	// use to get client count periodically
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				if client := reverseServer.GetTheNumberOfClient(); client > 0 {
					log.Printf("Current number of client %d", reverseServer.GetTheNumberOfClient())
				}
			case <-clientMonitorCancelSignal:
				ticker.Stop()
				return
			}

		}
	}()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	reverseServer.Shutdown(nil)
}

func createTestURLs() []ApplicationServerEnpoint {
	katoEx, _ := url.Parse("http://kato")
	kutoriEx, _ := url.Parse("http://kutori")
	superSimple, _ := url.Parse("http://localhost:8000")

	katoEndpoint := ApplicationServerEnpoint{
		Name:            "Kato",
		ExternalAddress: katoEx,
		LocalAddress:    superSimple,
	}

	kutoriEndpoint := ApplicationServerEnpoint{
		Name:            "Kutori",
		ExternalAddress: kutoriEx,
		LocalAddress:    superSimple,
	}

	return []ApplicationServerEnpoint{katoEndpoint, kutoriEndpoint}
}
