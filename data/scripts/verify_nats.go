package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	fmt.Printf("Connecting to NATS at %s...\n", natsURL)
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Subscribing to 'upload' subject...")
	sub, err := js.SubscribeSync("upload")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for messages (60s timeout)...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Timeout reached.")
			return
		default:
			msg, err := sub.NextMsg(1 * time.Second)
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				log.Fatal(err)
			}

			fmt.Printf("Received message: %s\n", string(msg.Data))
			msg.Ack()
		}
	}
}
