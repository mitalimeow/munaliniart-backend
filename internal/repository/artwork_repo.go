package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Artwork struct {
	ID         int       `json:"id"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type,omitempty"`
	ImageData  []byte    `json:"-"` // never send raw bytes in JSON listing
	UploadedAt time.Time `json:"uploaded_at"`
}

type ArtworkRepository struct {
	DB *pgxpool.Pool
}

func NewArtworkRepository(db *pgxpool.Pool) *ArtworkRepository {
	return &ArtworkRepository{DB: db}
}

// GetArtworks returns metadata for all uploaded artworks, newest first
func (r *ArtworkRepository) GetArtworks(ctx context.Context) ([]Artwork, error) {
	query := `SELECT id, filename, uploaded_at FROM art ORDER BY uploaded_at DESC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artworks []Artwork
	for rows.Next() {
		var a Artwork
		if err := rows.Scan(&a.ID, &a.Filename, &a.UploadedAt); err != nil {
			return nil, err
		}
		artworks = append(artworks, a)
	}
	return artworks, nil
}

// GetArtworkImage returns the raw bytes and mime type for serving
func (r *ArtworkRepository) GetArtworkImage(ctx context.Context, id int) ([]byte, string, error) {
	var data []byte
	var mimeType string
	query := `SELECT image_data, mime_type FROM art WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&data, &mimeType)
	return data, mimeType, err
}

// CreateArtwork inserts a new artwork record into the database
func (r *ArtworkRepository) CreateArtwork(ctx context.Context, filename, mimeType string, data []byte) (int, error) {
	var id int
	query := `INSERT INTO art (filename, mime_type, image_data) VALUES ($1, $2, $3) RETURNING id`
	err := r.DB.QueryRow(ctx, query, filename, mimeType, data).Scan(&id)
	return id, err
}

// DeleteArtwork removes an artwork record
func (r *ArtworkRepository) DeleteArtwork(ctx context.Context, id int) error {
	query := `DELETE FROM art WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	return err
}
