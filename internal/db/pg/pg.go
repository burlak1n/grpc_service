package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	database "grpc_service/internal/db"
	"grpc_service/internal/domain/models"

	_ "github.com/lib/pq"
)

type DB struct {
	data *sql.DB
}

func (db *DB) Stop() error {
	return db.data.Close()
}

// New creates a new instance of postgres
func New(storagePath string) (*DB, error) {
	const op = "db.pg.New"

	db, err := sql.Open("postgres", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &DB{
		data: db,
	}, nil
}

func (db *DB) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte,
) (uint32, error) {
	const op = "db.pg.SaveUser"

	stmt, err := db.data.Prepare("INSERT INTO users(email, pass hash) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		// TODO: обработка ошибок
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// ID последней созданной записи (не будет ли проблем с параллелизмом?)
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uint32(id), nil
}

func (db *DB) User(
	ctx context.Context,
	email string,
) (User models.User, err error) {
	const op = "db.pg.User"
	var emptyUser = models.User{}

	stmt, err := db.data.Prepare("SELECT id, email, pass_hash FROM users")
	if err != nil {
		return emptyUser, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, email)

	err = row.Scan(&User.ID, &User.Email, &User.PassHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return emptyUser, fmt.Errorf("%s: %w", op, database.ErrUserNotFound)
		}
		return emptyUser, fmt.Errorf("%s: %w", op, err)
	}

	return User, nil
}

func (db *DB) IsAdmin(
	ctx context.Context,
	userID uint32,
) (bool, error) {
	const op = "db.pg.IsAdmin"

	stmt, err := db.data.Prepare("SELECT is_admin FROM users WHERE id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, database.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// App returns app by id.
func (db *DB) App(
	ctx context.Context,
	appID uint32,
) (models.App, error) {
	const op = "db.pg.App"

	stmt, err := db.data.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, database.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
