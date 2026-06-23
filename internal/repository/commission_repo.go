package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CommissionTier struct {
	ID          int       `json:"id"`
	TierName    string    `json:"tier_name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	IsAvailable bool      `json:"is_available"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommissionRepository struct {
	DB *pgxpool.Pool
}

func NewCommissionRepository(db *pgxpool.Pool) *CommissionRepository {
	return &CommissionRepository{DB: db}
}

func (r *CommissionRepository) GetTiers(ctx context.Context) ([]CommissionTier, error) {
	query := `SELECT id, tier_name, price, description, is_available, updated_at FROM commissions ORDER BY price ASC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []CommissionTier
	for rows.Next() {
		var tier CommissionTier
		err := rows.Scan(&r.DB, &tier.TierName, &tier.Price, &tier.Description, &tier.IsAvailable, &tier.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, tier)
	}
	return tiers, nil
}

func (r *CommissionRepository) UpdateTier(ctx context.Context, id int, name string, price float64, desc string, avail bool) error {
	query := `UPDATE commissions SET tier_name = $1, price = $2, description = $3, is_available = $4, updated_at = NOW() WHERE id = $5`
	_, err := r.DB.Exec(ctx, query, name, price, desc, avail, id)
	return err
}
