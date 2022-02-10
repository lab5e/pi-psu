package main

import (
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/stianeikeland/go-rpio/v4"
)

var opt struct {
	Pin int `short:"p" default:"27" description:"pin to toggle for 1 second"`
}

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
		}
	}
}

func main() {
	err := rpio.Open()
	if err != nil {
		log.Fatalf("unable to open GPIO: %v", err)
	}
	defer rpio.Close()

	pin := rpio.Pin(opt.Pin)

	pin.High()
	time.Sleep(time.Second)
	pin.Low()
}
