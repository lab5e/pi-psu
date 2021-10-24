package spanlistener

import (
	"context"
	"log"
	"time"

	"github.com/lab5e/go-spanapi/v4"
	"github.com/lab5e/go-spanapi/v4/apitools"
)

// SpanListener implements a very simple listener for Span.
type SpanListener struct {
	config      Config
	dataChannel chan string
	ctx         context.Context
	cancel      context.CancelFunc
}

// Config parameters for SpanListener.
type Config struct {
	Token      string
	Collection string
	Device     string
}

const (
	queueLength    = 100
	enqueueTimeout = 10 * time.Millisecond
	reconnectDelay = 30 * time.Second
)

// New creates a new Span listener
func New(c Config) *SpanListener {
	ctx, cancel := context.WithCancel(context.Background())

	spanListener := &SpanListener{
		config:      c,
		dataChannel: make(chan string, queueLength),
		ctx:         ctx,
		cancel:      cancel,
	}

	go spanListener.readLoop()
	return spanListener
}

// Data channel from Span listener
func (s *SpanListener) Data() <-chan string {
	return s.dataChannel
}

// Close the span listener.
func (s *SpanListener) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *SpanListener) readLoop() {
	defer close(s.dataChannel)

	spanAPIConfig := spanapi.NewConfiguration()
	for {
		ds, err := apitools.NewDeviceDataStream(apitools.ContextWithAuth(s.config.Token), spanAPIConfig, s.config.Collection, s.config.Device)
		if err != nil {
			log.Printf("error connecting to: %v", err)
			log.Printf("sleeping for %s before reconnect", reconnectDelay)

			select {
			case <-time.After(reconnectDelay):
			case <-s.ctx.Done():
				return
			}

			continue
		}

		// reception loop
		for {
			msg, err := ds.Recv()
			if err != nil {
				log.Printf("error receiving message: %v", err)
				continue
			}

			select {
			case s.dataChannel <- *msg.Payload:

			case <-time.After(enqueueTimeout):
				log.Printf("queue full, dropping message")

			case <-s.ctx.Done():
				ds.Close()
				return
			}
		}
	}
}
