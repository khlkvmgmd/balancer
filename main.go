package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	servers         = []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002", "http://localhost:3003", "http://localhost:3004"}
	requestCounters = make(map[string]int)
	mu              sync.Mutex
)

func main() {

	for _, server := range servers {
		requestCounters[server] = 0
	}

	go logRequestCounters()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		server := selectServer()

		mu.Lock()
		requestCounters[server]++
		mu.Unlock()

		proxyRequest(w, r, server)
	})

	fmt.Println("Load balancer is running on :8080...")
	http.ListenAndServe(":8080", nil)
}

func selectServer() string {
	mu.Lock()
	defer mu.Unlock()
	selectedServer := servers[0]
	servers = append(servers[1:], servers[0])
	return selectedServer
}

func proxyRequest(w http.ResponseWriter, r *http.Request, server string) {
	resp, err := http.Get(server + r.URL.Path + "ping")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		w.Header()[name] = values
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func logRequestCounters() {
	for {
		time.Sleep(3 * time.Second)
		mu.Lock()
		for server, count := range requestCounters {
			fmt.Printf("Server %s: %d requests\n", server, count)
		}
		mu.Unlock()
	}
}
