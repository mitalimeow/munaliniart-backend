package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Artwork represents an uploaded image stored on disk + recorded in DB
type Artwork struct {
	ID        int       `json:"id"`
	Filename  string    `json:"filename"`
	Filepath  string    `json:"filepath"`
	CreatedAt time.Time `json:"created_at"`
}

type ArtworkRepository struct {
	DB *pgxpool.Pool
}

func NewArtworkRepository(db *pgxpool.Pool) *ArtworkRepository {
	return &ArtworkRepository{DB: db}
}

// GetArtworks returns all uploaded artworks, newest first
func (r *ArtworkRepository) GetArtworks(ctx context.Context) ([]Artwork, error) {
	query := `SELECT id, filename, filepath, created_at FROM artworks ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artworks []Artwork
	for rows.Next() {
		var a Artwork
		if err := rows.Scan(&a.ID, &a.Filename, &a.Filepath, &a.CreatedAt); err != nil {
			return nil, err
		}
		artworks = append(artworks, a)
	}
	return artworks, nil
}

// CreateArtwork inserts a new artwork record into the database
func (r *ArtworkRepository) CreateArtwork(ctx context.Context, filename, filepath string) (int, error) {
	var id int
	query := `INSERT INTO artworks (filename, filepath) VALUES ($1, $2) RETURNING id`
	err := r.DB.QueryRow(ctx, query, filename, filepath).Scan(&id)
	return id, err
}

// DeleteArtwork removes an artwork record and returns the filepath so the handler can delete the file
func (r *ArtworkRepository) DeleteArtwork(ctx context.Context, id int) (string, error) {
	var filepath string
	query := `DELETE FROM artworks WHERE id = $1 RETURNING filepath`
	err := r.DB.QueryRow(ctx, query, id).Scan(&filepath)
	return filepath, err
}
