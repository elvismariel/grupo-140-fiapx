package repository

import (
	"context"
	"video-processor-worker/internal/core/domain"
	"video-processor-worker/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVideoRepository struct {
	db *pgxpool.Pool
}

func NewPostgresVideoRepository(db *pgxpool.Pool) ports.VideoRepository {
	return &postgresVideoRepository{
		db: db,
	}
}

func (r *postgresVideoRepository) Update(ctx context.Context, video *domain.Video) error {
	query := `
		UPDATE videos
		SET status = $1, zip_path = $2, frame_count = $3, message = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query, video.Status, video.ZipPath, video.FrameCount, video.Message, video.ID).
		Scan(&video.UpdatedAt)
	return err
}

func (r *postgresVideoRepository) GetByID(ctx context.Context, id int64) (*domain.Video, error) {
	query := `SELECT id, user_id, filename, status, COALESCE(zip_path, ''), frame_count, COALESCE(message, ''), created_at, updated_at FROM videos WHERE id = $1`
	video := &domain.Video{}
	err := r.db.QueryRow(ctx, query, id).
		Scan(&video.ID, &video.UserID, &video.Filename, &video.Status, &video.ZipPath, &video.FrameCount, &video.Message, &video.CreatedAt, &video.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return video, err
}

func (r *postgresVideoRepository) GetPending(ctx context.Context) ([]domain.Video, error) {
	query := `SELECT id, user_id, filename, status, COALESCE(zip_path, ''), frame_count, COALESCE(message, ''), created_at, updated_at FROM videos WHERE status = 'PENDING' ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var v domain.Video
		err := rows.Scan(&v.ID, &v.UserID, &v.Filename, &v.Status, &v.ZipPath, &v.FrameCount, &v.Message, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}
