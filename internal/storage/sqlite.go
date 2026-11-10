package storage

import (
	"context"
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"

	"github.com/becomegloomy/go-tasks-api/internal/model"
)

var ErrNotFound = errors.New("task not found")

type Store interface {
	List(ctx context.Context) ([]model.Task, error)
	Get(ctx context.Context, id int64) (model.Task, error)
	Create(ctx context.Context, title string) (model.Task, error)
	Update(ctx context.Context, id int64, title string, done bool) (model.Task, error)
	Delete(ctx context.Context, id int64) error
	Close() error
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	done INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

func (s *SQLiteStore) List(ctx context.Context) ([]model.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, done, created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *SQLiteStore) Get(ctx context.Context, id int64) (model.Task, error) {
	var t model.Task
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = ?`, id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (s *SQLiteStore) Create(ctx context.Context, title string) (model.Task, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO tasks(title, done) VALUES(?, 0)`, title)
	if err != nil {
		return model.Task{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Task{}, err
	}
	return s.Get(ctx, id)
}

func (s *SQLiteStore) Update(ctx context.Context, id int64, title string, done bool) (model.Task, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE tasks SET title = ?, done = ? WHERE id = ?`, title, done, id)
	if err != nil {
		return model.Task{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.Task{}, ErrNotFound
	}
	return s.Get(ctx, id)
}

func (s *SQLiteStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
