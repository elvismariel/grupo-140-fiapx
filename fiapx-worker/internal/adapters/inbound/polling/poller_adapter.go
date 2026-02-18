package polling

import (
	"context"
	"log"
	"time"
	"video-processor-worker/internal/core/ports"
)

type PollerAdapter struct {
	repo    ports.VideoRepository
	handler func(ctx context.Context, videoID int64) error
}

func NewPollerAdapter(repo ports.VideoRepository, handler func(ctx context.Context, videoID int64) error) *PollerAdapter {
	return &PollerAdapter{
		repo:    repo,
		handler: handler,
	}
}

func (a *PollerAdapter) Start(ctx context.Context) {
	log.Println("🚀 Poller started, monitoring for pending videos (fallback)...")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("👋 Stopping poller...")
			return
		case <-ticker.C:
			videos, err := a.repo.GetPending(ctx)
			if err != nil {
				log.Printf("❌ Error polling videos: %v", err)
				continue
			}

			if len(videos) == 0 {
				continue
			}

			for _, v := range videos {
				log.Printf("🎬 Poller found video ID: %d (%s)", v.ID, v.Filename)
				if err := a.handler(ctx, v.ID); err != nil {
					log.Printf("❌ Error handling video %d from poller: %v", v.ID, err)
				}
			}
		}
	}
}
