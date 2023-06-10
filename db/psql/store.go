package psql

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/thunder33345/bookstore"
)

// sqlErrUniqueViolation is a constant used match sql code and generate more useful errors
const sqlErrUniqueViolation = "23505"

// sqlErrForeignKeyViolation is a constant used match sql code and generate more useful errors
const sqlErrForeignKeyViolation = "23503"

// sqlErrRaisedException is used to match errors raised in sql and replaced with more useful errors
const sqlErrRaisedException = "P0001"

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var name string
	err = s.db.GetContext(ctx, &name, `SELECT extname FROM pg_extension WHERE extname = 'pg_trgm' LIMIT 1`)
	if err != nil {
		return fmt.Errorf(`error selecting for "pg_trgm": %w`, err)
	}
	if name != "pg_trgm" {
		return fmt.Errorf(`error missing psql extension "pg_trgm", make sure it's enabled and installed'`)
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

// enrichPQError attempts to adds error type to a pq error
func enrichPQError(err error, resource string) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		//pass through: leave the error untouched if it's not a pq error
		return err
	}

	switch pqErr.Code {
	case sqlErrUniqueViolation:
		err = bookstore.NewDuplicateError(resource, err)
	case sqlErrForeignKeyViolation:
		switch pqErr.Constraint { //we use constraint to return more user-friendly errors
		case "fk_author":
			err = bookstore.NewInvalidDependencyError("author", err)
		case "fk_genre":
			err = bookstore.NewInvalidDependencyError("genre", err)
		}
	}
	return err
}

// enrichListPQError attempts to add error type on a listing request
// currently it only handles raised exception, which is raised in our query code
func enrichListPQError(err error, resource string) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		//pass through: leave the error untouched if it's not a pq error
		return err
	}

	switch pqErr.Code {
	case sqlErrRaisedException:
		err = bookstore.NewNonExistentIDError(resource, err)
	}
	return err
}

func enrichDeletePQError(err error, resource string) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		//pass through: leave the error untouched if it's not a pq error
		return err
	}

	switch pqErr.Code {
	case sqlErrForeignKeyViolation: //foreign key errors while deleting means something is being depended on
		err = bookstore.NewDependedError(resource, err)
	}
	return err
}

// checkAffectedRows is a helper function to simplify checking for affected row
// caller provides the result and noResErr to return, if the affected row is <=0
func checkAffectedRows(res sql.Result, noResErr error) error {
	//we try to check affected row
	rows, err := res.RowsAffected()
	if err != nil {
		//this shouldn't happen, but we account for it anyway
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if rows <= 0 {
		//if the affected rows is <=0 we return the supplied error
		return noResErr
	}
	//we return nil if everything is ok
	return nil
}
