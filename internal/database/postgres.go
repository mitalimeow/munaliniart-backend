package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB establishes a thread-safe connection pool to your PostgreSQL database
func ConnectDB(databaseURL string) (*pgxpool.Pool, error) {
	// 1. Create a background context with a 10-second timeout limit.
	// If your database is turned off, we don't want our Go application hanging forever.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Parse the configuration string from your .env file
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Printf("Unable to parse database URL connection string: %v\n", err)
		return nil, err
	}

	// 3. Optimize the connection pool metrics for a lightweight server
	config.MaxConns = 10                     // Allow up to 10 simultaneous database connections max
	config.MinConns = 2                      // Keep at least 2 connections open at all times
	config.MaxConnIdleTime = 5 * time.Minute // Close idle extra connections after 5 mins

	// 4. Instantiate the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Failed to create PostgreSQL connection pool: %v\n", err)
		return nil, err
	}

	// 5. Ping the database to ensure our credentials work and the database is actively responsive
	err = pool.Ping(ctx)
	if err != nil {
		log.Printf("Database ping failed. Connection could not be confirmed: %v\n", err)
		pool.Close()
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL! Connection pool is ready.")
	return pool, nil
}
