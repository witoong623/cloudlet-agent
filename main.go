package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/witoong623/restserver"
)

func main() {
	// Create demo applicatoin server addresses and start reverse proxy module
	applicationServers := createTestURLs()
	reverseServer, _ := NewMultiHostReverseProxyServer(applicationServers)
	go func() {
		reverseServer.StartReverseProxyServer()
	}()

	// Create ManagementCommunication instance which is a module that
	// manage all of communication with Management server
	managementCom := NewManagementCommunication("127.0.0.1", "127.0.0.1", "localhost")
	// REST server is used to host any service offer to Management server
	cloudletRESTServer := restserver.NewRESTServer(":6000")
	cloudletRESTServer.Handle("/info/currentclient", HandleManagementWorkloadQuery(managementCom))
	go func() {
		cloudletRESTServer.StartListening()
	}()

	// Poppulate Client Count
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				managementCom.ClientCountIncoming <- reverseServer.GetTheNumberOfClient()
			}
		}
	}()

	if err := managementCom.RegisterThisCloudlet("http://localhost:8000/cloudlet/register"); err != nil {
		log.Printf("cannot register this Cloudlet to Management server, Error %v", err.Error())
	} else {
		log.Println("successfully register to Management server")
	}

	/* clientMonitorCancelSignal := make(chan struct{}, 1)
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
	}() */

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan
	reverseServer.Shutdown(nil)
	cloudletRESTServer.StopListening(nil)
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
