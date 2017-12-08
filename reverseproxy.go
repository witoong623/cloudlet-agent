package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
)

type reverseProxyCtx struct {
	targets map[string]CountableReverseProxy
}

// ApplicationServerEnpoint is used to store information
// about internet address of service and local address
// of application server
type ApplicationServerEnpoint struct {
	Name            string
	ExternalAddress *url.URL
	LocalAddress    *url.URL
}

// ReverseProxyServer is a type that represents reverse proxy server
type ReverseProxyServer struct {
	httpServer      *http.Server
	reverseProxyCtx *reverseProxyCtx
}

// NewMultiHostReverseProxyServer creates new instance of ReverseProxyServer
// for proxy connection to application server given target URLs.
func NewMultiHostReverseProxyServer(targets []ApplicationServerEnpoint) (*ReverseProxyServer, error) {
	targetMap := make(map[string]CountableReverseProxy)
	for _, endpoint := range targets {
		crp := NewSingleHostCountableReverseProxy(endpoint.LocalAddress, true)
		targetMap[endpoint.ExternalAddress.Host] = *crp
	}
	reverseProxyCtx := &reverseProxyCtx{targets: targetMap}
	reverseproxyServer := &ReverseProxyServer{
		httpServer:      &http.Server{Addr: ":80"},
		reverseProxyCtx: reverseProxyCtx,
	}
	http.Handle("/", handleReverseProxy(reverseProxyCtx))

	return reverseproxyServer, nil
}

// StartReverseProxyServer will start and proxy all of clients connection
// Calling this method will get blocked.
func (rps *ReverseProxyServer) StartReverseProxyServer() error {
	err := rps.httpServer.ListenAndServe()
	return err
}

// Shutdown will shutdown reverse proxy after all request is served.
func (rps *ReverseProxyServer) Shutdown(ctx context.Context) error {
	return rps.httpServer.Shutdown(ctx)
}

// GetTheNumberOfClient returns total number of clients that connecting through this proxy server
func (rps *ReverseProxyServer) GetTheNumberOfClient() int32 {
	var total int32
	for _, reverseproxy := range rps.reverseProxyCtx.targets {
		total += reverseproxy.GetClientConnection()
	}

	return total
}

func handleReverseProxy(rpc *reverseProxyCtx) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reverseProxy, ok := rpc.targets[r.Host]; ok {
			reverseProxy.ServeHTTP(w, r)
		} else {
			log.Println("Server not found")
		}
	})
}
