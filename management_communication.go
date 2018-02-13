package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	mobyioutil "github.com/docker/docker/pkg/ioutils"
)

var httpClient http.Client

// ManagementCommunication encapsulates information about this node
type ManagementCommunication struct {
	Name                     string
	IPAddr                   string
	Domain                   string
	httpClient               *http.Client
	ClientCountIncoming      chan int32
	mutex                    sync.Mutex
	clientCountOutgoingChans map[int64]chan int32
}

type WorkloadStatusMessage struct {
	ClientCount int32
}

// NewManagementCommunication returns pointer to ManagementCommunication
// given name, ip and domain name of this Cloudlet
func NewManagementCommunication(name, ip, domain string) *ManagementCommunication {
	instance := &ManagementCommunication{
		httpClient:               &httpClient,
		Name:                     name,
		IPAddr:                   ip,
		Domain:                   domain,
		ClientCountIncoming:      make(chan int32),
		clientCountOutgoingChans: make(map[int64]chan int32),
	}

	go instance.broadcastClientCount()

	return instance
}

// RegisterThisCloudlet registers this Cloudlet with Management server
func (m *ManagementCommunication) RegisterThisCloudlet(registerLink string) error {
	formData := url.Values{}
	formData.Add("cloudlet-name", m.Name)
	formData.Add("cloudlet-ip", m.IPAddr)
	formData.Add("cloudlet-domain", m.Domain)
	res, err := m.httpClient.PostForm(registerLink, formData)
	if err != nil {
		return err
	}
	res.Body.Close()

	return nil
}

// RegisterService registers service that this cloudlet offers
func (m *ManagementCommunication) RegisterService(registerLink, sName, sDomain string) error {
	formData := url.Values{}
	formData.Add("cloudlet-name", m.Name)
	formData.Add("service-name", sName)
	formData.Add("service-domain", sDomain)
	res, err := m.httpClient.PostForm(registerLink, formData)
	if err != nil {
		return err
	}
	res.Body.Close()

	return nil
}

// HandleManagementWorkloadQuery handles HTTP request from management server
func HandleManagementWorkloadQuery(cm *ManagementCommunication) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		timestamp := time.Now().Unix()
		outChan := make(chan int32, 10)

		cm.mutex.Lock()
		cm.clientCountOutgoingChans[timestamp] = outChan
		cm.mutex.Unlock()
		defer func() {
			cm.mutex.Lock()
			delete(cm.clientCountOutgoingChans, timestamp)
			cm.mutex.Unlock()
		}()

		rCancelSignal := r.Context().Done()
		writeflusher := mobyioutil.NewWriteFlusher(w)
		defer writeflusher.Close()
		writeflusher.Flush()
		encoder := json.NewEncoder(writeflusher)
		var workloadMsg WorkloadStatusMessage

		for {
			select {
			case <-rCancelSignal:
				return
			default:
				workloadMsg = WorkloadStatusMessage{ClientCount: <-outChan}
				if err := encoder.Encode(&workloadMsg); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})
}

func (m *ManagementCommunication) broadcastClientCount() {
	LongRunGOWG.Add(1)
	defer LongRunGOWG.Done()
	for count := range m.ClientCountIncoming {
		m.mutex.Lock()
		for _, out := range m.clientCountOutgoingChans {
			out <- count
		}
		m.mutex.Unlock()
	}
}

// GetServerRequestingCount returns the number
// of client that are requesting workload count
func (m *ManagementCommunication) GetServerRequestingCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return len(m.clientCountOutgoingChans)
}
