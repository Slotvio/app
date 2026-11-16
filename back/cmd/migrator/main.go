package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/aleksandrboro/slotvio/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var migrationsPath, migrationsType string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsType, "migrations-type", "", "type of migrations: up or down")
	flag.Parse()

	if migrationsPath == "" {
		return errors.New("provide migrations path")
	}
	if migrationsType != "up" && migrationsType != "down" {
		return fmt.Errorf("invalid migrations type %q: want up or down", migrationsType)
	}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load optional .env: %w", err)
	}

	var dbCfg config.DBCfg
	if err := cleanenv.ReadEnv(&dbCfg); err != nil {
		return fmt.Errorf("read database environment: %w", err)
	}
	if dbCfg.User == "" || dbCfg.Password == "" || dbCfg.Host == "" || dbCfg.Port == "" || dbCfg.Name == "" {
		return errors.New("DB_USER, DB_PASS, DB_HOST, DB_PORT and DB_NAME must be set")
	}

	sourceURL := (&url.URL{Scheme: "file", Path: migrationsPath}).String()
	databaseURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(dbCfg.User, dbCfg.Password),
		Host:   net.JoinHostPort(dbCfg.Host, dbCfg.Port),
		Path:   "/" + dbCfg.Name,
	}

	query := databaseURL.Query()
	query.Set("sslmode", "disable")
	databaseURL.RawQuery = query.Encode()

	m, err := migrate.New(sourceURL, databaseURL.String())
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}
	defer m.Close()

	switch migrationsType {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	}
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no change")
		return nil
	}
	if err != nil {
		return fmt.Errorf("apply %s migrations: %w", migrationsType, err)
	}

	fmt.Printf("%s migrations applied successfully\n", migrationsType)
	return nil
}
