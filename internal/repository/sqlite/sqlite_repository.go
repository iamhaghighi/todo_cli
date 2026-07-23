package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"todo_cli/internal/domain"

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

func New(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	r := &Repository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Repository) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		done INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) GetAll(ctx context.Context) ([]domain.Todo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			title,
			done,
			created_at,
			updated_at
		FROM todos
	`)
	// ORDER BY created_at DESC
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []domain.Todo

	for rows.Next() {
		var (
			id        int64
			title     string
			doneInt   int
			createdAt string
			updatedAt string
		)

		if err := rows.Scan(&id, &title, &doneInt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		created, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		updated, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}

		todos = append(todos, domain.Todo{
			ID:        id,
			Title:     title,
			Done:      doneInt != 0,
			CreatedAt: created,
			UpdatedAt: updated,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *Repository) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO todos(
			title,
			done,
			created_at,
			updated_at
		)
		VALUES(?,?,?,?)
	`,
		todo.Title,
		boolToInt(todo.Done),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return domain.Todo{}, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return domain.Todo{}, err
	}

	created := domain.Todo{
		ID:        lastID,
		Title:     todo.Title,
		Done:      todo.Done,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return created, nil
}

func (r *Repository) Update(ctx context.Context, todo domain.Todo) error {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, `
		UPDATE todos
		SET
			title = ?,
			updated_at = ?
		WHERE id = ?
		`,
		todo.Title,
		now.Format(time.RFC3339Nano),
		todo.ID,
	)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return errors.New("todo not found")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM todos
		WHERE id = ?
	`, id)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return errors.New("todo not found")
	}
	return nil
}

func (r *Repository) Toggle(ctx context.Context, id int64, done bool) error {
	res, err := r.db.ExecContext(ctx, `
	UPDATE todos
	SET
		done = ?
	WHERE id = ?
	`,
		boolToInt(done),
		id,
	)

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return errors.New("todo not found")
	}
	return nil
}

func boolToInt(b bool) int {
	if b == true {
		return 1
	}
	return 0
}
