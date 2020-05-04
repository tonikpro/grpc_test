package test

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_mysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type Migration struct {
	Migrate *migrate.Migrate
}

func (m Migration) Up() (bool, error) {
	err := m.Migrate.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func (m Migration) Down() (bool, error) {
	err := m.Migrate.Down()
	if err != nil {
		return false, err
	}
	return true, nil
}

func NewMigration(dbConn *sqlx.DB, migrationsFolderLocation string) (*Migration, error) {
	pathToMigrate := fmt.Sprintf("file://%s", migrationsFolderLocation)

	driver, err := _mysql.WithInstance(dbConn.DB, &_mysql.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(pathToMigrate, "mysql", driver)
	if err != nil {
		return nil, err
	}
	return &Migration{Migrate: m}, nil
}
