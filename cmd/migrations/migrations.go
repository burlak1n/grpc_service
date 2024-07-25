package main

import (
	"errors"
	"flag"
	"fmt"
	plog "grpc_service/internal/slogger/prettyhandler"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/ilyakaznacheev/cleanenv"

	// Драйвер postgres
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type ConfigDatabase struct {
	Port           int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	Host           string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Name           string `yaml:"name" env:"POSTGRES_DB" env-default:"postgres"`
	User           string `yaml:"user" env:"POSTGRES_USER" env-default:"user"`
	Password       string `yaml:"password" env:"POSTGRES_PASSWORD"`
	MigrationsPath string `yaml:"password" env:"POSTGRES_MIGRATIONS_PATH"`
}

func Env(cfg *ConfigDatabase) (err error) {
	err = cleanenv.ReadConfig(".env", cfg)
	if err != nil {
		return err
	}
	return err
}

func Flags(dbPath, migrationsPath *string) {
	// user:password@host:port/dbname
	var dbPathFlag, migrationsPathFlag string

	flag.StringVar(&dbPathFlag, "db-path", "", "path to db")
	flag.StringVar(&migrationsPathFlag, "migrations-path", "", "path to migrations")
	// flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if dbPathFlag != "" {
		*dbPath = dbPathFlag
	}

	if migrationsPathFlag != "" {
		*migrationsPath = migrationsPathFlag
	}
}

// https://github.com/golang-migrate/migrate/tree/master/database/postgres
func main() {
	var cfg ConfigDatabase
	log := plog.SetupLogger("local")

	if err := Env(&cfg); err != nil {
		log.Error("Failed to init env", plog.Err(err))
		return
	}

	dbPath := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	Flags(&dbPath, &cfg.MigrationsPath)

	// setup database
	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		dbPath,
	)

	if err != nil {
		log.Error("Failed to create migration", plog.Err(err))
		return
	}
	old_version_uint, is_dirty, err := m.Version()
	old_version := int(old_version_uint)
	if err != nil {
		if err.Error() != "no migration" {
			log.Error("can't take current migration version", plog.Err(err))
			return
		}
	}
	if is_dirty {
		log.Error("current version is dirty", slog.Int("cur_version", old_version))
		return
	}
	log.Info("Starting migrations",
		slog.Any("config", cfg),
		slog.String("dbPath", dbPath),
		slog.Int(
			"cur_version",
			old_version,
		),
	)

	if err := m.Up(); err != nil {
		log.Warn("error with migration up", plog.Err(err))

		if errors.Is(err, migrate.ErrNoChange) {
			log.Error("no migrations to apply", plog.Err(err))
			return
		}
		new_version_uint, is_dirty, err_v := m.Version()
		new_version := int(new_version_uint)
		if err_v != nil {
			log.Error("can't take new migration version:", slog.Int("new_version", new_version), plog.Err(err_v))
			return
		}
		if is_dirty {
			log.Warn("new version is dirty. Forcing to previous version...")

			err = m.Force(old_version)

			if err != nil {
				log.Error("can't force to old version", slog.Int("new_version", new_version), plog.Err(err))
				return
			}
		}
	}
	new_version_uint, _, _ := m.Version()
	new_version := int(new_version_uint)

	log.Info("migrations applied successfully", slog.Int("new_version", new_version))
}
