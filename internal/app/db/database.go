package db

import (
	"bobot/internal/config"
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"log"
	"os"
)

func dbConnStr() string {
	cfg := config.PostgresConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("unable to parse PostgresConfig: %v", err)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresDbHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDb)
}

func NewConnection(ctx context.Context) (*pgxDB, error) {
	connStr := dbConnStr()
	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return NewDatabase(pool), nil
}

func MigrationUp() {
	db, err := goose.OpenDBWithDriver("postgres", dbConnStr())

	if err != nil {
		log.Fatal(err)
	}

	if err = goose.Up(db, os.Getenv("MIGRATION_FOLDER")); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

type DB interface {
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	GetPool(ctx context.Context) *pgxpool.Pool
}

type pgxDB struct {
	cluster *pgxpool.Pool
}

func NewDatabase(cluster *pgxpool.Pool) *pgxDB {
	return &pgxDB{cluster: cluster}
}

func (db pgxDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.cluster, dest, query, args...)
}

func (db pgxDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.cluster, dest, query, args...)
}

func (db pgxDB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.cluster.Exec(ctx, query, args...)
}

func (db pgxDB) ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.cluster.QueryRow(ctx, query, args...)
}

func (db pgxDB) GetPool(_ context.Context) *pgxpool.Pool {
	return db.cluster
}
