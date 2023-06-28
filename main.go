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
	addr string
	proxy *httputil.ReverseProxy
}


type LoadBalancer struct {
	port		string
	roundRobinCount		int
	servers		[]Server
}

func newSimpleServer(addr string) *simpleServer{
	serverUrl, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		 addr: addr,
		 proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer (port string, servers []Server) *LoadBalancer{
	return &LoadBalancer{
		port: port,
		roundRobinCount: 0,
		servers: servers,
	}
}

//handle error function 
func handleError(err error){
	if err != nil {
		fmt.Printf("error: %v\n" , err)
		os.Exit(1)
	}
}

func (s *simpleServer) Address() string{ return s.addr}

func (s *simpleServer) IsAlive() bool {return true}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request){
	s.proxy.ServeHTTP(rw, req)
}


//get the next available server method
func (lb *LoadBalancer) getNextAvailableServer() Server{
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]

	for !server.IsAlive(){
		lb.roundRobinCount ++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]

	}

	lb.roundRobinCount++
	return server
}

//server proxy method
func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request){
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwading request to address %q\n", targetServer)
	targetServer.Serve(rw, req)
}

//the main function 
func main(){
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.twitter.com"),
	}

	lb := *NewLoadBalancer("8000", servers)

	handleRedirect := func(rw http.ResponseWriter, req *http.Request){
		lb.serveProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Serving request at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":" + lb.port, nil)
}