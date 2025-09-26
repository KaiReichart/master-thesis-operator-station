package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kaireichart/master-thesis-operator-station/data_analysis"
	"github.com/kaireichart/master-thesis-operator-station/events"
	"github.com/kaireichart/master-thesis-operator-station/gps"
	"github.com/kaireichart/master-thesis-operator-station/mental_rotation"
	"github.com/kaireichart/master-thesis-operator-station/programs"
)

func init() {
	events.Init()
	gps.Init()
	programs.Init()
	mental_rotation.Init()
	data_analysis.Init()
}

func main() {
	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		if err := data_analysis.CloseMainDatabase(); err != nil {
			log.Printf("Error closing main database: %v", err)
		}
		os.Exit(0)
	}()

	// Serve static files
	http.Handle("/manifest.json", http.FileServer(http.Dir(".")))
	http.Handle("/icons/", http.StripPrefix("/icons/", http.FileServer(http.Dir("icons"))))
	http.HandleFunc("/", serveFrontend)

	events.SetupHandlers()
	gps.SetupHandlers()
	programs.SetupHandlers()
	mental_rotation.SetupHandlers()
	data_analysis.SetupHandlers()

	log.Printf("Server started at http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)
}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "overview.html")
}
