package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/witoong623/restserver"
)

// LongRunGOWG is used to wait any long run goroutine to exit before terminating program.
// Any code can Add and Done this WG.
var LongRunGOWG sync.WaitGroup

func main() {
	// read config file
	err := ReadConfigurationFile("config.json")
	if err != nil {
		log.Fatalln(err)
	}
	// Create demo applicatoin server addresses and start reverse proxy module
	applicationServers := createTestURLs()
	reverseServer, _ := NewMultiHostReverseProxyServer(applicationServers)
	go func() {
		reverseServer.StartReverseProxyServer()
	}()

	// Create ManagementCommunication instance which is a module that
	// manage all of communication with Management server
	// and create REST server to host service for cloudlet agent
	managementCom := NewManagementCommunication(Config.CloudletName, Config.CloudletIP, Config.CloudletDomain)
	cloudletRESTServer := restserver.NewRESTServer(":6000")
	cloudletRESTServer.Handle("/info/currentclient", HandleManagementWorkloadQuery(managementCom))
	go func() {
		cloudletRESTServer.StartListening()
	}()

	// Populate Client Count to middle value, reduces lock at reverse proxy
	popCancelChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				if managementCom.GetServerRequestingCount() == 0 {
					continue
				}
				managementCom.ClientCountIncoming <- reverseServer.GetTheNumberOfClient()
			case <-popCancelChan:
				close(managementCom.ClientCountIncoming)
				ticker.Stop()
				return
			}
		}
	}()

	if err := managementCom.RegisterThisCloudlet(Config.CloudletRegisterLink); err != nil {
		log.Printf("cannot register this cloudlet to management server, Error %v", err.Error())
	} else {
		log.Println("successfully register to management server")
	}
	fs := Config.Services[0]
	if err := managementCom.RegisterService(Config.ServiceRegisterLink, fs.Name, fs.Domain); err != nil {
		log.Printf("cannot register service %v with domain %v to management server", fs.Name, fs.Domain)
	} else {
		log.Printf("successfully register service %v to management server", fs.Name)
	}

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, syscall.SIGTERM, syscall.SIGINT)
	<-closeChan

	// Terminate everything
	popCancelChan <- struct{}{}
	reverseServer.Shutdown(nil)
	cloudletRESTServer.StopListening(nil)
}

func createTestURLs() []ApplicationServerEnpoint {
	ocrExternal, _ := url.Parse("http://ocr.aliveplex.net")
	ocrInternal, _ := url.Parse("http://localhost:8080")

	ocrEndPoint := ApplicationServerEnpoint{
		Name:            "OCR",
		ExternalAddress: ocrExternal,
		LocalAddress:    ocrInternal,
	}

	return []ApplicationServerEnpoint{ocrEndPoint}
}
