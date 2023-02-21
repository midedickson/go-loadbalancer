package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

func newSimpleServer(address string) *simpleServer {
	serveUrl, err := url.Parse(address)

	handleErr(err)
	return &simpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serveUrl),
	}
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string {
	return s.address
}

func (s *simpleServer) IsAlive() bool {
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwaring request to addresses: %v\n", targetServer.Address())
	targetServer.Serve(rw, r)

}

func main() {
	servers := []Server{
		newSimpleServer("https://facebook.com"),
		newSimpleServer("https://bing.com"),
		newSimpleServer("https://duckduckgo.com"),
	}
	lb := NewLoadBalancer("8080", servers)
	handleRedirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serveProxy(rw, r)

	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Server is starting on the localhost at : %s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
