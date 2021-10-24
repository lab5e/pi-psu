package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"github.com/lab5e/pi-psu/pkg/spanlistener"
	"github.com/stianeikeland/go-rpio/v4"
)

var opt struct {
	HTTPAddr              string        `long:"addr" default:":8080" description:"HTTP interface listen address"`
	Token                 string        `long:"token" env:"PI_PSU_TOKEN" description:"Span API token" required:"yes"`
	Collection            string        `long:"collection" env:"PI_PSU_COLLECTION" description:"Span collection id" required:"yes"`
	Device                string        `long:"device" env:"PI_PSU_DEVICE" description:"Span device ID" required:"yes"`
	MessageTimeout        time.Duration `long:"msg-timeout" default:"2m" description:"time after last message seen from device to when we power cycle"`
	MinimumRebootInterval time.Duration `long:"minimum-reboot-interval" default:"5m" description:"minimum time between reboots"`
	GPIONum               int           `long:"gpio-pin" default:"27" description:"GPIO pin for relay"`
	GPIOHoldTime          time.Duration `long:"gpio-hold" default:"3s" description:"how long we turn off power"`
}

var state struct {
	LastRebootTime           time.Time `json:"lastRebootTime"`
	LastMessageTime          time.Time `json:"lastMessageTime"`
	LastRebootNumSecondsAgo  int       `json:"lastRebootNumSecondsAgo"`
	LastMessageNumSecondsAgo int       `json:"lastMessageNumSecondsAgo"`
	NumReboots               int       `json:"numReboots"`
}

var (
	pin rpio.Pin
)

func init() {
	p := flags.NewParser(&opt, flags.Default)
	_, err := p.Parse()
	if err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(0)
		}
	}
}

func main() {
	// Open GPIO
	err := rpio.Open()
	if err != nil {
		log.Fatalf("error opening GPIO: %v", err)
	}
	defer rpio.Close()

	pin = rpio.Pin(opt.GPIONum)

	// Fire up web interface
	mux := mux.NewRouter()
	mux.HandleFunc("/", statusHandler).Methods("GET")
	mux.HandleFunc("/reset", resetHandler).Methods("GET")

	server := http.Server{
		Addr:              opt.HTTPAddr,
		Handler:           mux,
		ReadTimeout:       20 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Printf("http server error: %v", err)
		}
	}()

	// Fire up listener
	listener := spanlistener.New(spanlistener.Config{
		Token:      opt.Token,
		Collection: opt.Collection,
		Device:     opt.Device,
	})
	defer listener.Close()

	for {
		select {
		case <-listener.Data():
			state.LastMessageTime = time.Now()
			log.Print("got packet")

		case <-time.After(opt.MessageTimeout):
			if time.Since(state.LastRebootTime) > opt.MinimumRebootInterval {
				state.LastRebootTime = time.Now()
				log.Printf("REBOOT")
				cycleRelay()
			} else {
				log.Printf("to soon to reboot")
			}
		}
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Update state
	state.LastRebootNumSecondsAgo = int(time.Since(state.LastRebootTime).Seconds())
	state.LastMessageNumSecondsAgo = int(time.Since(state.LastMessageTime).Seconds())

	jsonData, err := json.MarshalIndent(&state, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("error rendering state: %v", err), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	go cycleRelay()
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	fmt.Fprintf(w, "power cycle initiated, complete in %s", opt.GPIOHoldTime)
}

// cycle relay
func cycleRelay() {
	pin.High()
	time.Sleep(opt.GPIOHoldTime)
	pin.Low()
	state.NumReboots++
}
