package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"video-processor/internal/core/ports"

	"github.com/nats-io/nats.go"
)

type NatsAdapter struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

type uploadEvent struct {
	VideoID  int64  `json:"video_id"`
	Filename string `json:"filename"`
}

func NewNatsAdapter(url string) (ports.EventPublisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("error getting JetStream context: %w", err)
	}

	// Ensure stream exists
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "video",
		Subjects: []string{"upload"},
	})
	if err != nil {
		// If stream already exists, we might get an error or it might be fine depending on config.
		// AddStream is idempotent if config matches, but let's be safe.
		// In newer versions AddStream is fine if subjects overlap but names match.
		log.Printf("Note: Stream 'video' creation result: %v", err)
	}

	return &NatsAdapter{
		nc: nc,
		js: js,
	}, nil
}

func (a *NatsAdapter) PublishUploadEvent(videoID int64, filename string) error {
	event := uploadEvent{
		VideoID:  videoID,
		Filename: filename,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	_, err = a.js.Publish("upload", data)
	if err != nil {
		return fmt.Errorf("error publishing to NATS: %w", err)
	}

	log.Printf("Event published to NATS: video_id=%d, filename=%s", videoID, filename)
	return nil
}
