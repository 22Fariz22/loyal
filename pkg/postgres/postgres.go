package postgres

import (
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New -.
func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	createTables(pg.Pool)

	return pg, nil
}

// Close -.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func createTables(pool *pgxpool.Pool) (*Postgres, error) {
	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users(
		user_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		login VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		balance_total INT DEFAULT 0,
		withdraw_total INT DEFAULT 0
);

		CREATE TABLE IF NOT EXISTS orders(
		order_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		user_id INT,
		number VARCHAR(120) UNIQUE NOT NULL,
		order_status VARCHAR(15),
		accrual INT DEFAULT 0,
		uploaded_at timestamp NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_user
			FOREIGN KEY(user_id)
			REFERENCES users(user_id) 
);
		CREATE TABLE IF NOT EXISTS history(
		history_id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		user_id INT,
		number VARCHAR(120) UNIQUE NOT NULL,
		sum INT DEFAULT 0,
		processed_at timestamp NOT NULL DEFAULT NOW() ,
		CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(user_id)
		--CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(order_id)	
);
`)
	if err != nil {
		log.Printf("Unable to create table: %v\n", err)
		return nil, err
	}
	return nil, err
}
