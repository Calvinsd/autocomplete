package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Calvisd/autocomplete/search"
)

type RecommendationResponse struct {
	Recommendations []string `json:"recommendations"`
}

type SearchResponse struct {
	Found bool `json:"found"`
}

func main() {

	var interruptSignal = make(chan bool)

	go handleShutdowns(interruptSignal)

	// Initialize datastore
	dataStore := search.NewDataStore()
	dataStore.InitializeDataStore()

	//Web Interface

	const port string = ":8080"

	var serverShutdownSignal = make(chan bool)

	mux := http.NewServeMux()

	mux.HandleFunc("/", homePage)

	mux.HandleFunc("/search/recommendations", recommendData(dataStore))

	mux.HandleFunc("/search", searchData(dataStore))

	// server config
	server := http.Server{
		Addr:         port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      mux,
	}

	go func() {
		fmt.Println("Started server on port:", port)
		err := server.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting the server", err)
		}

		fmt.Println("Shutting down server...")

		serverShutdownSignal <- true

	}()

	<-interruptSignal

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}

	<-serverShutdownSignal
}

// home page reqest handler
func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 PAGE NOT FOUND"))
		return
	}
	http.ServeFile(w, r, "./index.html")
}

// search request handler
func searchData(dataStore *search.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["q"]; r.Method != http.MethodGet || !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))

			return
		}

		result := dataStore.Search(r.URL.Query()["q"][0])

		response := SearchResponse{
			Found: result.Found,
		}

		serializedResponse, err := json.Marshal(response)

		if err != nil {
			log.Fatal("Error serializing response", err)
		}

		w.WriteHeader(http.StatusOK)

		w.Write(serializedResponse)
	}
}

// Recommendation request handler
func recommendData(dataStore *search.DataStore) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["q"]; r.Method != http.MethodGet || !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))

			return
		}

		result := dataStore.Search(r.URL.Query()["q"][0])

		response := RecommendationResponse{
			Recommendations: result.Recommendations,
		}

		serializedResponse, err := json.Marshal(response)

		if err != nil {
			log.Fatal("Error serializing response", err)
		}

		w.WriteHeader(http.StatusOK)

		w.Write(serializedResponse)

	}

}

// listens for shutdown signals
func handleShutdowns(done chan<- bool) {
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGSEGV)

	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			log.Println("Encountered os interrupt")
			done <- true
		case syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGSEGV:
			log.Println("Received linux signel")
			done <- true
		}
	}()
}
