package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/enson89/user-service-go/internal/model"
	"github.com/jmoiron/sqlx"
)

// UserRepository wraps a sqlx.DB to manage users.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository constructs a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user record (inside a transaction) and returns the generated ID.
func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const query = `
        INSERT INTO users (email, password_hash, role)
        VALUES ($1, $2, $3)
        RETURNING id
    `
	// tx.GetContext will scan the returned id into u.ID
	if err := tx.GetContext(ctx, &u.ID, query, u.Email, u.PasswordHash, u.Role); err != nil {
		return err
	}

	return tx.Commit()
}

// GetByEmail fetches a user by email. Returns (nil, nil) if not found.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	const query = `
        SELECT id, email, password_hash, role
        FROM users
        WHERE email = $1
    `
	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetByID retrieves a user by their primary key. Returns (nil, nil) if not found.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	const query = `
        SELECT id, email, password_hash, role
        FROM users
        WHERE id = $1
    `
	err := r.db.GetContext(ctx, &u, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Delete removes a user by ID. Returns an error if no rows were affected.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("no user found to delete")
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const q = `
      UPDATE users
         SET name = $1, updated_at = NOW()
       WHERE id = $2
    `
	res, err := tx.ExecContext(ctx, q, u.Name, u.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	// e.g. insert into audit_logs (uncomment if you have this table)
	// _, err = tx.ExecContext(ctx,
	//     `INSERT INTO audit_logs(user_id, action, ts) VALUES ($1,'update_name', NOW())`,
	//     u.ID,
	// )
	// if err != nil {
	//     return err
	// }

	return tx.Commit()
}
