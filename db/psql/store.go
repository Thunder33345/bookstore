package psql

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/thunder33345/bookstore"
)

// sqlErrUniqueViolation is a constant used match sql code and generate more useful errors
const sqlErrUniqueViolation = "23505"

// sqlErrRestrictViolation is a constant used match sql code and generate more useful errors
const sqlErrRestrictViolation = "23001"

// sqlErrForeignKeyViolation is a constant used match sql code and generate more useful errors
const sqlErrForeignKeyViolation = "23503"

type Store struct {
	db *sqlx.DB
	//sqlDb is the standard sql.DB instance, used for migration
	sqlDb *sql.DB
}

// New creates a new store instance
// it expects connection string containing the necessary details to connect to a psql db
// see more at https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
func New(connStr string) (*Store, error) {
	//first we create the plain sql.db and store it, this is necessary for migrations
	sqlDb, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	//we wrap the sql.db in sqlx, this is what we will normally use
	db := sqlx.NewDb(sqlDb, "postgres")

	return &Store{
		sqlDb: sqlDb,
		db:    db,
	}, nil
}

// Init initializes store by testing for connectivity and perform migrations
func (s *Store) Init() error {
	err := s.db.Ping()
	if err != nil {
		return fmt.Errorf("error pinging: %w", err)
	}
	err = s.migrate()
	if err != nil {
		return fmt.Errorf("error migrating: %w", err)
	}
	return nil
}

// migrationFS stores the sql migrations files via embed
//
//go:embed migrations/*.sql
var migrationFS embed.FS

// migrate attempts to run migrations on the database to sync the database state with application state
func (s *Store) migrate() error {
	sqlDriver, err := postgres.WithInstance(s.sqlDb, &postgres.Config{
		MigrationsTable:       "",
		MigrationsTableQuoted: false,
		MultiStatementEnabled: false,
		DatabaseName:          "",
		SchemaName:            "",
		StatementTimeout:      0,
		MultiStatementMaxSize: 0,
	})
	if err != nil {
		return fmt.Errorf("error creating psql driver: %w", err)
	}

	srcDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("error creating fs driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", sqlDriver)
	if err != nil {
		return fmt.Errorf("error creating migration: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migration: %w", err)
	}
	return nil
}
